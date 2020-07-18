package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/hedzr/errors.v2"
	"net"
)

type Obj struct {
	listener    net.Listener
	connections []*connectionObj
	closeErr    error
	exitCh      chan struct{}
	newConnFunc NewConnectionFunc
}

func newServerObj(listener net.Listener) (s *Obj) {
	s = &Obj{
		listener:    listener,
		connections: nil,
		closeErr:    nil,
		exitCh:      make(chan struct{}),
		newConnFunc: newConnObj,
	}
	return
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
	//		fmt.Printf("Some connection error: %s\n", err)
	//		continue
	//	}
	//}

	ctx := context.Background()
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
				logrus.Errorf("can't accept a connection: %v", e)
			}
			return e
		}

		var co ConnectionObj
		co = s.newConnection(conn)
		go co.HandleConnection(ctx)
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

type NewConnectionFunc func(s *Obj, conn net.Conn) ConnectionObj

var ErrServerClosed = errors.New("server closed")
