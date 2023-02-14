/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hedzr/go-socketlib/tcp/tls"
)

const (
	DefaultBufferSize = 4096
)

type OnTcpServerCreateReadWriter func(ss *Server, conn net.Conn, tsConnected time.Time) (in io.Reader, out io.Writer)
type OnTcpServerConnectedWithClient func(ss *Server, conn net.Conn)
type OnTcpServerDisconnectedWithClient func(ss *Server, conn net.Conn, reader io.Reader)
type OnTcpServerProcessFunc func(buf []byte, in io.Reader, out io.Writer) (nn int, err error)
type OnTcpServerListening func(ss *Server, l net.Listener)

// Processor 代表在reader处理读取到到报文的同时会立即进行报文的处理。
//
// Reader负责从读取的报文数据块中按照协议进行分包，切分成功
// 的包（Packet）将被Processor所处理以完成业务逻辑。
//
// 如果Reader并未实现Processor接口，Server将会把识别到到
// 包交给 OnTcpServerProcessFunc 去处理。
type Processor interface {
	// Process implements Processor interface to announce that i will process the incoming data in Read().
	Process(buf []byte, in io.Reader, out io.Writer) (nn int, err error)
}

type Server struct {
	host        string
	port        int
	l           net.Listener
	done        chan struct{}
	wg          sync.WaitGroup
	exitingFlag bool

	base

	CmdrTlsConfig *tls.CmdrTlsConfig

	bufferSize                        int
	onTcpProcess                      OnTcpServerProcessFunc
	onTcpServerCreateReadWriter       OnTcpServerCreateReadWriter
	onTcpServerConnectedWithClient    OnTcpServerConnectedWithClient
	onTcpServerDisconnectedWithClient OnTcpServerDisconnectedWithClient
	onTcpServerListening              OnTcpServerListening
}

func newServer(addr string, opts ...ServerOpt) *Server {
	s := &Server{
		base:       newBase(nil),
		done:       make(chan struct{}),
		bufferSize: DefaultBufferSize,
	}

	var port string
	var err error
	s.host, port, err = net.SplitHostPort(addr)
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
	return s
}

func (s *Server) defaultCreateReadWriter(ss *Server, conn net.Conn, tsConnected time.Time) (in io.Reader, out io.Writer) {
	in = bufio.NewReader(conn)
	out = bufio.NewWriter(conn)
	// o = conn // bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	return
}

func (s *Server) Start() (err error) {
	s.exitingFlag = false

	if s.done == nil {
		s.done = make(chan struct{})
	}

	if s.onTcpProcess == nil {
		s.onTcpProcess = s.defaultProcess
	}

	if s.onTcpServerCreateReadWriter == nil {
		s.onTcpServerCreateReadWriter = s.defaultCreateReadWriter
	}

	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	// var l net.Listener
	s.l, err = net.Listen("tcp", addr)
	if err != nil {
		s.Errorf("error listening: addr=%v: %v", addr, err)
		return // os.Exit(1)
	}
	// NOTE NOTE NOTE: we ignore s.InitTlsConfigFromConfigFile() NOW because it has been done by via tcp.NewCmdrTlsConfig()
	if s.CmdrTlsConfig.IsCertValid() {
		s.l, err = s.CmdrTlsConfig.NewTlsListener(s.l)
		if err != nil {
			s.Errorf("error listening over TLS: addr=%v: %v", addr, err)
			return // os.Exit(1)
		}
		s.Debugf("A tcp server listening on %v (over TLS)", addr)
	} else {
		// defer l.Close()
		s.Debugf("A tcp server listening on %v", addr)
	}
	// s.debug("  > listening on %v", addr)
	// s.debug("    > listening on %v", addr)

	// s.wg.Add(2)
	// go s.handleWrite(s.conn, &s.wg)
	// go s.handleRead(s.conn, &s.wg)
	// s.wg.Wait()

	go s.runLoop(s.l, s.done)
	return
}

func (s *Server) Stop() {
	_ = s.Close()
}

func (s *Server) Close() (err error) {
	s.exitingFlag = true

	if s.l != nil {
		if err = s.l.Close(); err != nil {
			s.Errorf("closing s.listener: %v", err)
		}
	}

	if s.done != nil {
		close(s.done)
		s.done = nil
	}

	// if s.conn != nil {
	// 	if err := s.conn.Close(); err != nil {
	// 		s.Wrong(err, "closing s.conn")
	// 	}
	// 	s.conn = nil
	// }

	return
}

func (s *Server) runLoop(l net.Listener, done <-chan struct{}) {
	// timer := time.NewTicker(10 * time.Second)
	// defer func() {
	// 	timer.Stop()
	// 	s.Debug("[tcp][server] runLoop goroutine exited.")
	// }()

	if s.onTcpServerListening != nil {
		s.onTcpServerListening(s, l)
	}

	for {
		// select {
		// case <-done:
		// 	return
		// case tick := <-timer.C:
		// 	s.Debug("tick at %v", tick)

		conn, err := l.Accept()
		if err != nil {
			if s.exitingFlag {
				return
			}
			if neterr, ok := err.(net.Error); ok && (neterr.Temporary() || neterr.Timeout()) {
				s.Warnf("network error (temporary, or timeout), sleep 5ms and retry...: %v", neterr)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			s.Errorf("error accepting: %v", err)
			time.Sleep(5 * time.Millisecond)
			continue // os.Exit(1)
		}

		ts := time.Now().UTC()
		// logs an incoming message
		s.Debugf("received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())
		// Handle connections in a new goroutine.
		go s.handleRequest(conn, ts, done)
		// }
	}
}

func (s *Server) handleRequest(conn net.Conn, tsConnected time.Time, done <-chan struct{}) {
	var reader io.Reader
	var writer io.Writer
	defer func() {
		if err := conn.Close(); err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				s.Tracef("conn(from %v) closed by others.", conn.RemoteAddr())
			} else {
				s.Warnf("conn.close failed: %v", err)
			}
		}
		if s.onTcpServerDisconnectedWithClient != nil {
			s.onTcpServerDisconnectedWithClient(s, conn, reader)
		}
	}()

	if s.onTcpServerConnectedWithClient != nil {
		s.onTcpServerConnectedWithClient(s, conn)
	}

	// ctx, cancel := context.WithCancel(context.Background())
	// reader := bufio.NewReader(conn)
	// writer := bufio.NewWriter(conn)
	reader, writer = s.onTcpServerCreateReadWriter(s, conn, tsConnected)

	// ctxHolder, hasProcess := reader.(mqtt.Contextual)
	cidHolder, _ := reader.(interface{ GetClientID() string })
	_, hasProcess := reader.(Processor)

	buf := make([]byte, s.bufferSize)
	var nn int
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				s.Debugf("♦︎ conn(from: %v) read i/o eof found. closing '%v'", conn.RemoteAddr(), cidHolder.GetClientID())
			} else {
				if n > 0 {
					_, _ = s.onTcpProcess(buf[:n], reader, writer)
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					s.Tracef("♦︎ conn(from %v) closed by others.", conn.RemoteAddr())
				} else {
					s.Errorf("♦︎︎ conn(from: %v) reader.read(buf) failed. closing '%v': %v", conn.RemoteAddr(), cidHolder.GetClientID(), err)
				}
			}
			return
		} else if n == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if hasProcess {
			// 如果 reader 对象也实现了 Processor 接口，则继续下一次读等待；
			// 此时代表着 reader 对象会在 Read() 的同时进行解码处理，所以
			// 由 WithServerOnProcessFunc() 提供的 onTcpProcess 回调函数将
			// 被忽略。
			continue
		}

		// 反之，则由 WithServerOnProcessFunc() 提供的 onTcpProcess 回调函数
		// 进行解码操作。

		// s.Debug("onTcpProcess processing %v bytes (%v, '%v')", nn, buf[:nn], string(buf[:nn]))
		nn, err = s.onTcpProcess(buf[:n], reader, writer)
		if err != nil {
			s.Errorf("onTcpProcess(buf, wr) failed. conn(from: %v), nn=%v. closing '%v': %v", conn.RemoteAddr(), nn, cidHolder.GetClientID(), err)
			return
		}
		// s.Debug("onTcpProcess processed %v bytes (%v, '%v')", nn, buf[:nn], string(buf[:nn]))

		// if err := Copyd(done, conn, conn); err != nil {
		// 	s.Wrong(err, "io.copy failed")
		// }
	}
}

func (s *Server) defaultProcess(buf []byte, in io.Reader, out io.Writer) (nn int, err error) {
	// nn, err = out.Write(buf)
	return
}

// func (s *Server) handleWrite(conn net.Conn, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	for i := 10; i > 0; i-- {
// 		_, err := conn.Write([]byte("hello " + strconv.Itoa(i) + "\r\n"))
// 		if err != nil {
// 			s.wrong(err, "error to send message in i=%v", i)
// 			break
// 		}
// 	}
// }
//
// func (s *Server) handleRead(conn net.Conn, wg *sync.WaitGroup) {
// 	defer wg.Done()
// 	reader := bufio.NewReader(conn)
// 	for i := 1; i <= 10; i++ {
// 		line, err := reader.ReadString(byte('\n'))
// 		if err != nil {
// 			s.wrong(err, "error to read message in i=%v ", i)
// 			return
// 		}
// 		fmt.Print(line)
// 	}
// }
