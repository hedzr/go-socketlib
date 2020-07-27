package server

import (
	"bufio"
	"context"
	"fmt"
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
	closeErr  error
	//exitCh    chan struct{}
	//logger    logx.Logger
}

func newConnObj(ctx context.Context, serverObj *Obj, conn net.Conn) (s ConnectionObj) {
	s = &connectionObj{
		serverObj: serverObj,
		conn:      conn,
		wrCh:      make(chan []byte, 256),
		//exitCh:    make(chan struct{}),
		//logger:    serverObj.logger,
	}
	return
}

func (s *connectionObj) Close() {
	if s.conn != nil {
		s.closeErr = s.conn.Close()
		s.conn = nil
	}
	close(s.wrCh)
	//close(s.exitCh)
}

func (s *connectionObj) HandleConnection(ctx context.Context) {
	s.serverObj.Printf("Client connected from " + s.RemoteAddrString())
	defer func() {
		s.serverObj.Printf("Client at " + s.RemoteAddrString() + " disconnected.")
	}()

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
	message := string(msg)
	s.serverObj.Tracef("> %v", message)

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

func (s *connectionObj) handleWriteRequests(ctx context.Context) {
	for {
		select {
		case msg := <-s.wrCh:
			s.doWrite(msg)
		case <-ctx.Done():
			return
		case <-ctx.Done():
			// If the request gets cancelled, log it
			// to STDERR
			s.serverObj.Errorf("request cancelled")
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
			s.serverObj.Errorf("Write message failed: %v (%v bytes written)", err, n)
		}
	}
}

func (s *connectionObj) RemoteAddrString() string {
	return s.conn.RemoteAddr().String()
}
