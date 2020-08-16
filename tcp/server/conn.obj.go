package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
	"net"
	"sync/atomic"
	"time"
)

type Connection interface {
	base.Conn
	base.CachedTCPWriter

	// HandleConnection is used by serverObj
	HandleConnection(ctx context.Context)

	//// WriteString send the string to the writing queue
	//WriteString(message string)
	//// Write send the buffer to the writing queue
	//Write(message []byte)
}

type connectionObj struct {
	serverObj *Obj
	uid       uint64
	conn      net.Conn
	wrCh      chan []byte
	closeErr  error
	//exitCh    chan struct{}
	//logger    logx.Logger
}

func newConnObj(ctx context.Context, serverObj *Obj, conn net.Conn) (s Connection) {
	co := &connectionObj{
		serverObj: serverObj,
		uid:       atomic.AddUint64(&serverObj.uidConn, 1),
		conn:      conn,
		wrCh:      make(chan []byte, 256),
		//exitCh:    make(chan struct{}),
		//logger:    serverObj.logger,
	}
	s = co
	return
}

func (s *connectionObj) Logger() log.Logger {
	return s.serverObj
}

func (s *connectionObj) Close() {
	if s.conn != nil {
		if s.serverObj.protocolInterceptor != nil {
			s.serverObj.protocolInterceptor.OnClosing(s, 0)
		}
		s.closeErr = s.conn.Close()
		s.conn = nil
	}
	close(s.wrCh)
	//close(s.exitCh)
	if s.serverObj.protocolInterceptor != nil {
		s.serverObj.protocolInterceptor.OnClosed(s, 0)
	}
}

func (s *connectionObj) HandleConnection(ctx context.Context) {
	s.serverObj.Debugf("[#%d] Client connected from %q", s.uid, s.RemoteAddrString())
	defer func() {
		s.serverObj.Debugf("[#%d] Client at %q disconnected.", s.uid, s.RemoteAddrString())
	}()

	if s.serverObj.protocolInterceptor != nil {
		s.serverObj.protocolInterceptor.OnConnected(ctx, s)
	}

	go s.handleWriteRequests(ctx)

	scanner := bufio.NewScanner(s.conn)
	for {
		ok := scanner.Scan()
		if !ok {
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}

		s.handleMessage(ctx, scanner.Bytes())
	}
}

func (s *connectionObj) handleMessage(ctx context.Context, msg []byte) {

	if s.serverObj.protocolInterceptor != nil {
		if processed, err := s.serverObj.protocolInterceptor.OnReading(ctx, s, msg); processed {
			return
		} else if err != nil {
			s.serverObj.Errorf("[#%d] error occurs on intercepting reading bytes: %v", s.uid, err)
			return
		}
	}

	message := string(msg)
	s.serverObj.Tracef("> [#%d] %v", s.uid, message)

	if len(message) > 0 && message[0] == '/' {
		switch {
		case message == "/time":
			resp := "It is " + time.Now().String() + "\n"
			fmt.Print("< " + resp)
			s.WriteString(resp)

		case message == "/quit":
			fmt.Println("Quitting.")
			s.WriteString("I'm shutting down now.\n")
			fmt.Println("< " + "%quit%")
			s.WriteString("%quit%\n")
			//os.Exit(0)
			//s.Close()
			s.serverObj.RequestShutdown()

		default:
			s.WriteString("Unrecognized command.\n")
		}
	}
}

func (s *connectionObj) handleWriteRequests(ctx context.Context) {
	for {
		select {
		case msg := <-s.wrCh:
			s.doWrite(ctx, msg)
		case <-ctx.Done():
			// If the request gets cancelled, log it
			// to STDERR
			s.serverObj.Debugf("[#%d] request cancelled", s.uid)
			return
		}
	}
}

func (s *connectionObj) WriteString(message string) {
	s.wrCh <- []byte(message)
}

func (s *connectionObj) Write(message []byte) {
	s.wrCh <- message
}

func (s *connectionObj) doWrite(ctx context.Context, msg []byte) {
	if s.conn != nil {

		if s.serverObj.protocolInterceptor != nil {
			if processed, err := s.serverObj.protocolInterceptor.OnWriting(ctx, s, msg); processed {
				return
			} else if err != nil {
				s.serverObj.Errorf("[#%d] error occurs on intercepting writing bytes: %v", s.uid, err)
				return
			}
		}

		var err error
		var n int
		err = s.conn.SetWriteDeadline(time.Now().Add(s.serverObj.WriteTimeout))
		if err != nil {
			s.serverObj.Errorf("[#%d] error set writing deadline: %v", s.uid, err)
			return
		}
		n, err = s.conn.Write(msg)
		if err != nil {
			s.serverObj.Errorf("[#%d] Write message failed: %v (%v bytes written)", s.uid, err, n)
		}
	}
}

func (s *connectionObj) RawWrite(ctx context.Context, msg []byte) (n int, err error) {
	if s.conn != nil {
		err = s.conn.SetWriteDeadline(time.Now().Add(s.serverObj.WriteTimeout))
		if err != nil {
			s.serverObj.Errorf("[#%d] error set writing deadline: %v", s.uid, err)
			return
		}
		n, err = s.conn.Write(msg)
	}
	return
}

func (s *connectionObj) RemoteAddrString() string {
	if s.conn != nil {
		return s.conn.RemoteAddr().String()
	}
	return ">?<"
}

func (s *connectionObj) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}
