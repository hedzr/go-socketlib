package server

import "net"

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

func (s *serverObj) Close() {
	if s.listener != nil {
		s.closeErr = s.listener.Close()
		s.listener = nil
	}

	for _, c := range s.connections {
		c.Close()
	}
}

func (s *serverObj) Accept() (co *connObj, err error) {
	var conn net.Conn
	conn, err = s.listener.Accept()
	if err != nil {
		return
	}
	co = newConnObj(conn)
	go co.handleConnection()
	return
}
