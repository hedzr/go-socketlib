package server

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tcp/udp"
	"github.com/hedzr/log"
	"net"
	"strings"
	"time"
)

type Obj struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	log.Logger
	listener            net.Listener
	udpConn             *udp.Obj
	connections         []*connectionObj
	closeErr            error
	exitCh              chan struct{}
	pfs                 *pidFileStruct
	newConnFunc         NewConnectionFunc
	protocolInterceptor protocol.Interceptor
	prefix              string
	uidConn             uint64
	netType             string
}

type NewConnectionFunc func(ctx context.Context, serverObj *Obj, conn net.Conn) Connection

func newServerObj(logger log.Logger) (s *Obj) {
	s = &Obj{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Logger:       logger,
		connections:  nil,
		closeErr:     nil,
		exitCh:       make(chan struct{}),
		newConnFunc:  newConnObj,
		netType:      "tcp",
	}
	//if s.Logger == nil {
	//	s.Logger = sugar.New("debug", false, true)
	//}
	return
}

func (s *Obj) ProtocolInterceptor() protocol.Interceptor {
	return s.protocolInterceptor
}

func (s *Obj) SetProtocolInterceptor(pi protocol.Interceptor) {
	s.protocolInterceptor = pi
}

func (s *Obj) ListenTo(listener net.Listener) {
	s.listener = listener
}

func (s *Obj) RequestShutdown() {
	close(s.exitCh)
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}
	if s.udpConn != nil {
		s.closeErr = s.udpConn.Close()
		s.udpConn = nil
	}
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
}

func (s *Obj) isUDP() bool {
	return strings.HasPrefix(s.netType, "udp")
}

func (s *Obj) createUDPListener(config *base.Config) (err error) {
	if s.udpConn == nil {
		s.udpConn = udp.NewUdpObj(s, nil, nil)
	}
	err = s.udpConn.Create(s.netType, config)
	return
}

func (s *Obj) createListener(config *base.Config) (tlsEnabled bool, err error) {
	// var listener net.Listener
	// var tlsEnabled bool
	s.listener, tlsEnabled, err = s.serverBuildListener(config.Addr, config.PrefixInConfigFile, config.PrefixInCommandLine)
	//if err != nil {
	//	s.Fatalf("build listener failed: %v", err)
	//}
	return
}

func (s *Obj) serverBuildListener(addr, prefixInConfigFile, prefixInCommandLine string) (listener net.Listener, tls bool, err error) {
	var tlsListener net.Listener

	listener, err = net.Listen(s.netType, addr)
	if err != nil {
		s.Fatalf("error: %v", err)
	}

	ctcPrefix := prefixInConfigFile + ".tls"
	ctc := tls2.NewCmdrTlsConfig(ctcPrefix, prefixInCommandLine)
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
	defer s.Close()
	//for {
	//	_, err := s.Accept()
	//	if err != nil {
	//		s.logger.Printf("Some connection error: %s\n", err)
	//		continue
	//	}
	//}

	// baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	if s.protocolInterceptor != nil {
		s.protocolInterceptor.OnServerReady(ctx, s)
		defer func() { s.protocolInterceptor.OnServerClosed(s) }()
	}

	switch s.isUDP() {
	case true:
		err = s.udpConn.Serve(ctx)
		if err != nil {
			s.Errorf("UDP serve failed: %v", err)
		}

	default:
		for {
			conn, e := s.listener.Accept()
			if e != nil {
				select {
				case <-s.exitCh:
					return ErrServerClosed
				case <-ctx.Done():
					return ErrServerClosed
				default:
				}

				if ne, ok := e.(net.Error); ok && ne.Temporary() {
					// handle the error
					s.Errorf("can't accept a connection: %v", e)
				}
				return e
			}

			var co Connection
			co = s.newConnection(ctx, conn)
			go co.HandleConnection(ctx)
			//c := srv.newConn(rw)
			//c.setState(c.rwc, StateNew) // before Serve can return
			//go c.serve(ctx)
		}
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
