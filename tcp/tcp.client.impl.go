/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"bufio"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/log"
	"github.com/hedzr/log/trace"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Client struct {
	host   string
	port   int
	conn   net.Conn
	done   chan struct{}
	wg     sync.WaitGroup
	closed int32

	base
	CmdrTlsConfig *tls.CmdrTlsConfig

	readBufferSize int
	verbose        bool

	connectedCh       chan net.Conn
	sendCh            chan []byte
	onTcpProcess      OnTcpProcessFunc
	onTcpConnected    OnTcpConnectedFunc
	onTcpDisconnected OnTcpDisconnectedFunc
}

type OnTcpConnectedFunc func(c *Client, conn net.Conn)
type OnTcpProcessFunc func(buf []byte, in *bufio.Reader, out *bufio.Writer) (nn int, err error)
type OnTcpDisconnectedFunc func(c *Client)

func NewClient(addr string, opts ...ClientOpt) *Client {
	return newClient(addr, opts...)
}

// func WithClientVerbose(verbose bool) ClientOpt {
// 	return func(client *Client) {
// 		client.verbose = verbose
// 	}
// }

func WithClientTlsConfig(s *tls.CmdrTlsConfig) ClientOpt {
	return func(client *Client) {
		client.CmdrTlsConfig = s
	}
}

func WithClientReadBufferSize(size int) ClientOpt {
	return func(client *Client) {
		client.readBufferSize = size
	}
}

func WithClientOnProcessFunc(fn OnTcpProcessFunc) ClientOpt {
	return func(client *Client) {
		client.onTcpProcess = fn
	}
}

func WithClientOnConnectedFunc(fn OnTcpConnectedFunc) ClientOpt {
	return func(client *Client) {
		client.onTcpConnected = fn
	}
}

func WithClientOnDisconnectedFunc(fn OnTcpDisconnectedFunc) ClientOpt {
	return func(client *Client) {
		client.onTcpDisconnected = fn
	}
}

//func WithClientLoggerConfig(config *log.LoggerConfig) ClientOpt {
//	return func(client *Client) {
//		client.Logger = build.New(config)
//	}
//}

func WithClientLogger(l log.Logger) ClientOpt {
	return func(client *Client) {
		client.Logger = l
	}
}

func newClient(addr string, opts ...ClientOpt) *Client {
	s := &Client{
		base:           newBase(nil),
		done:           make(chan struct{}),
		connectedCh:    make(chan net.Conn),
		sendCh:         make(chan []byte),
		readBufferSize: 4096,
	}

	var port string
	var err error
	s.host, port, err = net.SplitHostPort(addr)
	// s.wrong(err, "can't split addr to host & port")
	// s.wrong(err, "can't split addr to host & port")
	// s.wrong(err, "can't split addr to host & port")
	if err != nil {
		s.Errorf("can't split addr to host & port: %v", err)
		return nil
	}
	s.port, err = strconv.Atoi(port)
	if err != nil {
		s.Errorf("can't parse port to integer: %v", err)
		return nil
	}

	for _, opt := range opts {
		opt(s)
	}

	if err = s.run(); err != nil {
		s.Errorf("can't run(): 5v", err)
	}
	return s
}

func (s *Client) run() (err error) {
	if s.done == nil {
		s.done = make(chan struct{})
	}

	if s.sendCh == nil {
		s.sendCh = make(chan []byte)
	}

	if s.onTcpProcess == nil {
		s.onTcpProcess = s.defaultOnRead
	}

	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))

	go s.runLoop(s.done)

	var c net.Conn
	c, err = s.CmdrTlsConfig.Dial("tcp", addr)
	// s.conn, err = net.Dial("tcp", addr)
	if err != nil {
		s.Errorf("[tcp][client] error connecting to %v: %v", addr, err)
		s.Close()
		return // os.Exit(1)
	}
	s.conn = c
	s.Debugf("➠ [tcp][client] connected to %v", addr)
	// defer conn.Close()

	s.wg.Add(1)
	// go s.handleWrite(s.conn, &s.wg)
	go s.handleRead(s.conn, &s.wg)
	// s.wg.Wait()

	s.connectedCh <- s.conn

	// s.debug("[tcp][client] end of client looper")
	return
}

func (s *Client) IsClosed() bool {
	c := atomic.LoadInt32(&s.closed)
	return c == 1
}

func (s *Client) Close() {
	if s.done != nil {
		close(s.done)
		s.done = nil
	}

	s.closeConn()
}

func (s *Client) closeConn() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		if s.conn != nil {
			if err := s.conn.Close(); err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					s.Tracef("s.conn closed by others.")
				} else {
					s.Errorf("closing s.conn: %v", err)
				}

				s.conn = nil
			}
		}
	}
}

func (s *Client) runLoop(done <-chan struct{}) {
	// timer := time.NewTicker(10 * time.Second)
	defer func() {
		// timer.Stop()
		s.closeConn()
		s.Tracef("➠ [tcp][client] runLoop goroutine exited.")
	}()

	for {
		select {
		case <-done:
			return
		// case tick := <-timer.C:
		// 	s.Trace("tick at %v", tick)

		case c := <-s.connectedCh:
			if s.onTcpConnected != nil {
				s.onTcpConnected(s, c)
			}
		case data := <-s.sendCh:
			s.write_(data)
		}
	}
}

func (s *Client) Send(data []byte) {
	if s.IsClosed() {
		return
	}
	s.sendCh <- data
}

func (s *Client) write_(data []byte) {
	if data != nil {
		_, err := s.conn.Write(data)
		if err != nil {
			s.Errorf("error to send message: %v", err)
		} else if trace.IsEnabled() {
			s.Tracef("   -> TCP.W: % x", data)
		}
	}
}

// func (s *Client) handleWrite(conn net.Conn, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for i := 10; i > 0; i-- {
// 		_, err := conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))
// 		if err != nil {
// 			s.wrong(err, "error to send message in i=%v", i)
// 			break
// 		}
// 	}
// }

func (s *Client) defaultOnRead(p []byte, in *bufio.Reader, out *bufio.Writer) (n int, err error) {
	s.Debugf("read: %v", p)
	return 0, nil
}

func (s *Client) handleRead(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		if s.onTcpDisconnected != nil {
			s.onTcpDisconnected(s)
		}
		wg.Done()
	}()

	var nProcessed, n int
	var err error
	verbose := cmdr.GetBoolR("verbose")
	o := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	buf := make([]byte, s.readBufferSize)
	// conn.SetReadDeadline(time.Now().Add(5*time.Seconds))
	for {
		// err = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		// if err != nil {
		// 	s.Wrong(err, "SetReadDeadline failed")
		// 	break
		// }
		n, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				if n > 0 {
					s.Warnf("   tcp: EOF reached with %v bytes", n)
				}
				if s.IsClosed() {
					s.Debugf("    tcp: EOF reached. socket broken or closed")
					break
				}
				s.Debugf("    tcp: EOF reached. cancel reading...")
				err = connCheck(conn)
				time.Sleep(300 * time.Millisecond)
				break // can't recovery from this point, exit and close socket right now
			} else if e, ok := err.(net.Error); ok && e.Timeout() {
				continue
			} else if strings.Contains(err.Error(), "use of closed network connection") {
				s.Tracef("conn(from %v) closed by others.", conn.RemoteAddr())
			} else if strings.Contains(err.Error(), "connection reset by peer") {
				s.Tracef("conn(from %v) closed by peer.", conn.RemoteAddr())
			} else {
				s.Errorf("   tcp: read failed. reason: %v", err)
			}
			break
		} else if n == 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		vBuf := buf[:n]
		s.Tracef("   <- TCP.R [%v]: % x", verbose, vBuf)

		if nProcessed, err = s.onTcpProcess(vBuf, nil, o.Writer); err != nil {
			s.Errorf("   onTcpProcess returns failed: %v", err)
		}
		s.Tracef("   onTcpProcess processed %v bytes", nProcessed)
	}
}
