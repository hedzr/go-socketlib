package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tcp/udp"
)

const CTX_SERVER_OBJECT_KEY = "server-object"
const CTX_CONN_KEY = "conn"

type Obj struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	log.Logger

	listener            net.Listener
	udpConn             *udp.Obj
	connections         []*connectionObj
	closeErr            error
	pfs                 base.PidFile
	newConnFunc         NewConnectionFunc
	protocolInterceptor protocol.Interceptor
	prefix              string
	uidConn             uint64
	netType             string
	config              *base.Config
	// tlsConfigInitializer tls2.Initializer
}

type NewConnectionFunc func(ctx context.Context, serverObj *Obj, conn net.Conn) Connection

func newServerObj(config *base.Config) (s *Obj) {
	s = &Obj{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Logger:       nil,
		connections:  nil,
		closeErr:     nil,
		newConnFunc:  newConnObj,
		netType:      defaultNetType,
		config:       config,
	}
	// if s.Logger == nil {
	//	s.Logger = sugar.New("debug", false, true)
	// }
	return
}

func (s *Obj) BaseUri() string {
	return s.config.UriBase
}

func (s *Obj) Config() *base.Config {
	return s.config
}

func (s *Obj) SetLogger(l log.Logger) {
	s.Logger = l
}

func (s *Obj) WithLogger(logger log.Logger) *Obj {
	s.SetLogger(logger)
	return s
}

// func (s *Obj) WithTlsConfigInitializer(fn func(config *tls2.CmdrTlsConfig)) *Obj {
//	s.tlsConfigInitializer = fn
//	return s
// }

func (s *Obj) ProtocolInterceptor() protocol.Interceptor {
	return s.protocolInterceptor
}

func (s *Obj) SetProtocolInterceptor(pi protocol.Interceptor) {
	s.protocolInterceptor = pi
}

func (s *Obj) WithProtocolInterceptor(pi protocol.Interceptor) *Obj {
	s.protocolInterceptor = pi
	return s
}

func (s *Obj) ListenTo(listener net.Listener) {
	s.listener = listener
}

func (s *Obj) Close() {
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}
	if s.udpConn != nil {
		s.closeErr = s.udpConn.Close()
		s.udpConn = nil
	}

	for _, c := range s.connections {
		c.Close()
	}

	if s.pfs != nil {
		s.pfs.Destroy()
		s.pfs = nil
	}

	close(globalDoneCh) // *Obj server has completed its cleanup actions now
}

func (s *Obj) isUDP() bool {
	return strings.HasPrefix(s.netType, "udp")
}

func (s *Obj) createUDPListener(baseCtx context.Context, config *base.Config) (err error) {
	if s.udpConn == nil {
		s.udpConn = udp.New(s, nil, nil)
	}
	err = s.udpConn.Create(baseCtx, s.netType, config)
	return
}

func (s *Obj) createListener(baseCtx context.Context) (tlsEnabled bool, err error) {
	s.listener, tlsEnabled, err = s.serverBuildListener(baseCtx)
	return
}

func (s *Obj) serverBuildListener(baseCtx context.Context) (listener net.Listener, tls bool, err error) {
	var tlsListener net.Listener

	listener, err = net.Listen(s.netType, s.config.Addr)
	if err != nil {
		s.Fatalf("error: %v", err)
	}

	var ctc *tls2.CmdrTlsConfig
	if s.config.TlsConfigInitializer != nil {
		ctc = tls2.NewTlsConfig(s.config.TlsConfigInitializer)
	} else {
		ctcPrefix := s.config.PrefixInConfigFile + ".tls"
		ctc = tls2.NewCmdrTlsConfig(ctcPrefix, s.config.PrefixInCommandLine)
	}

	// s.Debugf("%v", ctc)
	if ctc.Enabled {
		tlsListener, err = ctc.NewTlsListener(listener)
		if err != nil {
			s.Fatalf("error: %v", err)
		}
	}

	if tlsListener != nil {
		listener = tlsListener
		tls = true
	}
	return
}

func (s *Obj) Serve(baseCtx context.Context) (err error) {
	// for {
	//	_, err := s.Accept()
	//	if err != nil {
	//		s.logger.Printf("Some connection error: %s\n", err)
	//		continue
	//	}
	// }

	// baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer func() {
		fmt.Println()
		s.Debugf("...Serve() ended.")
		cancel()
		s.Close()
	}()

	err = executors[s.isUDP()](s, ctx)
	return
}

var executors = map[bool]func(s *Obj, ctx context.Context) (err error){
	true:  udpExecutor,
	false: tcpExecutor,
}

func udpExecutor(s *Obj, ctx context.Context) (err error) {
	ctx = context.WithValue(ctx, CTX_CONN_KEY, s.udpConn)
	if s.protocolInterceptor != nil {
		defer func() {
			s.protocolInterceptor.OnServerClosed(s)
		}()
		s.protocolInterceptor.OnServerReady(ctx, s)
	}

	err = s.udpConn.Serve(ctx)
	if err != nil {
		s.Errorf("UDP serve failed: %v", err)
	}
	return
}

func tcpExecutor(s *Obj, ctx context.Context) (err error) {
	for {
		conn, e := s.listener.Accept()
		s.Tracef("...listener.Accept: err=%v", err)

		ctx = context.WithValue(ctx, CTX_CONN_KEY, conn)
		if s.protocolInterceptor != nil {
			defer func() {
				s.protocolInterceptor.OnServerClosed(s)
			}()
			s.protocolInterceptor.OnServerReady(ctx, s)
		}

		select {
		case <-globalExitCh:
			s.Debugf("...global exiting")
			return ErrServerClosed
		case <-ctx.Done():
			return ErrServerClosed
		default:
		}

		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				// handle the error
				s.Errorf("can't accept a connection: %v", e)
			}
			err = e
			break
		}

		var co Connection
		co = s.newConnection(ctx, conn)
		go co.HandleConnection(ctx)
		// c := srv.newConn(rw)
		// c.setState(c.rwc, StateNew) // before Serve can return
		// go c.serve(ctx)
	}
	return
}

func (s *Obj) newConnection(ctx context.Context, conn net.Conn) (co Connection) {
	co = s.newConnFunc(ctx, s, conn)
	return
}

func (s *Obj) SetNewConnectionFunc(fn NewConnectionFunc) {
	s.newConnFunc = fn
	return
}

func (s *Obj) RequestShutdown() {
	if s.listener != nil {
		close(globalExitCh)
	}
	s.Debugf("closing...")
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}
	if s.udpConn != nil {
		s.closeErr = s.udpConn.Close()
		s.udpConn = nil
	}
}

// Shutdown will close down the server gracefully
func Shutdown(serverObj *Obj) {
	serverObj.RequestShutdown()
	time.Sleep(5 * time.Millisecond)
	<-globalDoneCh
}

var globalExitCh = make(chan struct{})
var globalDoneCh = make(chan struct{})
