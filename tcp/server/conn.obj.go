package server

import (
	"bufio"
	"context"
	"fmt"
	"github.com/hedzr/logex"
	"net"
	"time"
)

type Connection interface {
	Logger() logex.Logger

	Close()
	// RawWrite does write through the internal net.Conn
	RawWrite(ctx context.Context, message []byte) (n int, err error)

	// HandleConnection is used by serverObj
	HandleConnection(ctx context.Context)

	// WriteString send the string to the writing queue
	WriteString(message string)
	// Write send the buffer to the writing queue
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

func newConnObj(ctx context.Context, serverObj *Obj, conn net.Conn) (s Connection) {
	s = &connectionObj{
		serverObj: serverObj,
		conn:      conn,
		wrCh:      make(chan []byte, 256),
		//exitCh:    make(chan struct{}),
		//logger:    serverObj.logger,
	}
	return
}

func (s *connectionObj) Logger() logex.Logger {
	return s.serverObj
}

func (s *connectionObj) Close() {
	if s.conn != nil {
		if s.serverObj.protocolInterceptor != nil {
			s.serverObj.protocolInterceptor.OnClosing(s)
		}
		s.closeErr = s.conn.Close()
		s.conn = nil
	}
	close(s.wrCh)
	//close(s.exitCh)
	if s.serverObj.protocolInterceptor != nil {
		s.serverObj.protocolInterceptor.OnClosed(s)
	}
}

func (s *connectionObj) HandleConnection(ctx context.Context) {
	s.serverObj.Debugf("Client connected from " + s.RemoteAddrString())
	defer func() {
		s.serverObj.Debugf("Client at " + s.RemoteAddrString() + " disconnected.")
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
			s.serverObj.Errorf("error occurs on intercepting reading bytes: %v", err)
			return
		}
	}

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
			s.doWrite(ctx, msg)
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

func (s *connectionObj) doWrite(ctx context.Context, msg []byte) {
	if s.conn != nil {

		if s.serverObj.protocolInterceptor != nil {
			if processed, err := s.serverObj.protocolInterceptor.OnWriting(ctx, s, msg); processed {
				return
			} else if err != nil {
				s.serverObj.Errorf("error occurs on intercepting writing bytes: %v", err)
				return
			}
		}

		n, err := s.conn.Write(msg)
		if err != nil {
			s.serverObj.Errorf("Write message failed: %v (%v bytes written)", err, n)
		}
	}
}

func (s *connectionObj) RawWrite(ctx context.Context, msg []byte) (n int, err error) {
	if s.conn != nil {
		n, err = s.conn.Write(msg)
	}
	return
}

func (s *connectionObj) RemoteAddrString() string {
	return s.conn.RemoteAddr().String()
}
