package server

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"time"
)

type connObj struct {
	conn     net.Conn
	wrCh     chan []byte
	exitCh   chan struct{}
	closeErr error
}

func newConnObj(conn net.Conn) (s *connObj) {
	s = &connObj{
		conn:   conn,
		wrCh:   make(chan []byte, 256),
		exitCh: make(chan struct{}),
	}
	return
}

func (s *connObj) Close() {
	if s.conn != nil {
		s.closeErr = s.conn.Close()
		s.conn = nil
	}
	close(s.wrCh)
	close(s.exitCh)
}

func (s *connObj) RemoteAddrString() string {
	return s.conn.RemoteAddr().String()
}

func (s *connObj) handleConnection() {
	fmt.Println("Client connected from " + s.RemoteAddrString())

	go s.handleWriteRequests()

	scanner := bufio.NewScanner(s.conn)
	for {
		ok := scanner.Scan()
		if !ok {
			break
		}

		s.handleMessage(scanner.Bytes())
	}

	fmt.Println("Client at " + s.RemoteAddrString() + " disconnected.")
}

func (s *connObj) handleMessage(msg []byte) {
	message := string(msg)
	fmt.Println("> " + message)

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
			os.Exit(0)

		default:
			s.WriteString("Unrecognized command.\n")
		}
	}
}

func (s *connObj) WriteString(message string) {
	s.wrCh <- []byte(message)
}

func (s *connObj) Write(message []byte) {
	s.wrCh <- message
}

func (s *connObj) doWrite(message []byte) {
	if s.conn != nil {
		n, err := s.conn.Write(message)
		if err != nil {
			logrus.Errorf("Write message failed: %v (%v bytes written)", err, n)
		}
	}
}

func (s *connObj) handleWriteRequests() {
	for {
		select {
		case msg := <-s.wrCh:
			s.doWrite(msg)
		case <-s.exitCh:
			return
		}
	}
}
