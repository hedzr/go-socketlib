package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

type ConnectionObj interface {
	Close()
	HandleConnection(ctx context.Context)
	WriteString(message string)
	Write(message []byte)
}

type connectionObj struct {
	serverObj *Obj
	conn      net.Conn
	wrCh      chan []byte
	exitCh    chan struct{}
	closeErr  error
}

func newConnObj(serverObj *Obj, conn net.Conn) (s *connectionObj) {
	s = &connectionObj{
		serverObj: serverObj,
		conn:      conn,
		wrCh:      make(chan []byte, 256),
		exitCh:    make(chan struct{}),
	}
	return
}

func (s *connectionObj) Close() {
	if s.conn != nil {
		s.closeErr = s.conn.Close()
		s.conn = nil
	}
	close(s.wrCh)
	close(s.exitCh)
}

func (s *connectionObj) HandleConnection(ctx context.Context) {
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

func (s *connectionObj) handleMessage(msg []byte) {
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
			//os.Exit(0)
			s.Close()
			s.serverObj.RequestShutdown()

		default:
			s.WriteString("Unrecognized command.\n")
		}
	}
}

func (s *connectionObj) handleWriteRequests() {
	for {
		select {
		case msg := <-s.wrCh:
			s.doWrite(msg)
		case <-s.exitCh:
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

func (s *connectionObj) doWrite(message []byte) {
	if s.conn != nil {
		n, err := s.conn.Write(message)
		if err != nil {
			logrus.Errorf("Write message failed: %v (%v bytes written)", err, n)
		}
	}
}

func (s *connectionObj) RemoteAddrString() string {
	return s.conn.RemoteAddr().String()
}
