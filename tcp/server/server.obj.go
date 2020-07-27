package server

import (
	"context"
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/logx/zap/sugar"
	"gopkg.in/hedzr/errors.v2"
	"net"
)

type Obj struct {
	listener    net.Listener
	connections []*connectionObj
	closeErr    error
	exitCh      chan struct{}
	newConnFunc NewConnectionFunc
	logger      logex.Logger
}

type NewConnectionFunc func(serverObj *Obj, conn net.Conn) ConnectionObj

func newServerObj(logger logex.Logger) (s *Obj) {
	s = &Obj{
		//listener:    listener,
		connections: nil,
		closeErr:    nil,
		exitCh:      make(chan struct{}),
		newConnFunc: newConnObj,
		logger:      logger,
	}
	if s.logger == nil {
		s.logger = sugar.New("debug", false, true)
	}
	return
}

func (s *Obj) Listen(listener net.Listener) {
	s.listener = listener
}

func (s *Obj) RequestShutdown() {
	close(s.exitCh)
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}
}

func (s *Obj) Close() {
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}

	for _, c := range s.connections {
		c.Close()
	}
}

func (s *Obj) Serve() (err error) {
	defer s.Close()
	//for {
	//	_, err := s.Accept()
	//	if err != nil {
	//		s.logger.Printf("Some connection error: %s\n", err)
	//		continue
	//	}
	//}

	baseCtx := context.Background()

	for {
		conn, e := s.listener.Accept()
		if e != nil {
			select {
			case <-s.exitCh:
				return ErrServerClosed
			default:
			}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				// handle the error
				s.logger.Errorf("can't accept a connection: %v", e)
			}
			return e
		}

		var co ConnectionObj
		co = s.newConnection(conn)
		go co.HandleConnection(baseCtx)
		//c := srv.newConn(rw)
		//c.setState(c.rwc, StateNew) // before Serve can return
		//go c.serve(ctx)
	}
}

func (s *Obj) newConnection(conn net.Conn) (co ConnectionObj) {
	co = s.newConnFunc(s, conn)
	return
}

func (s *Obj) SetNewConnectionFunc(fn NewConnectionFunc) {
	s.newConnFunc = fn
	return
}

var ErrServerClosed = errors.New("server closed")
