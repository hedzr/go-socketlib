package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"gopkg.in/hedzr/errors.v2"
	"net"
)

type serverObj struct {
	listener    net.Listener
	connections []*connObj
	closeErr    error
	exitCh      chan struct{}
}

func newServerObj(listener net.Listener) (s *serverObj) {
	s = &serverObj{
		listener:    listener,
		connections: nil,
		closeErr:    nil,
		exitCh:      make(chan struct{}),
	}
	return
}

func (s *serverObj) RequestShutdown() {
	close(s.exitCh)
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}
}

func (s *serverObj) Close() {
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}

	for _, c := range s.connections {
		c.Close()
	}
}

func (s *serverObj) Serve() (err error) {
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

		var co *connObj
		co = newConnObj(s, conn)
		go co.handleConnection(ctx)
		//c := srv.newConn(rw)
		//c.setState(c.rwc, StateNew) // before Serve can return
		//go c.serve(ctx)
	}
}

//func (s *serverObj) Accept() (co *connObj, err error) {
//	var conn net.Conn
//	conn, err = s.listener.Accept()
//	if err != nil {
//		return
//	}
//	co = newConnObj(conn)
//	go co.handleConnection()
//	return
//}

var ErrServerClosed = errors.New("server closed")
