package net

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/go-socketlib/net/api"
)

const defaultBufferSize = 4096

func NewServer(addr string, opts ...ServerOpt) *serverWrap {
	// var host, port string
	// var portN int
	// var err error
	// host, port, err = net.SplitHostPort(addr)
	// if err != nil {
	// 	var b = newBaseS()
	// 	b.error("can't split addr to host:port", "err", err, "addr", addr)
	// 	return nil
	// }
	// portN, err = strconv.Atoi(port)
	// if err != nil {
	// 	(&baseS{}).error("can't parse port to integer", "err", err, "port", port)
	// 	return nil
	// }

	s := &serverWrap{
		network:     "tcp",
		address:     addr,
		bufferSize:  defaultBufferSize,
		connections: make(map[*connS]bool),
		baseS:       newBaseS(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type Server interface {
	Listen(ctx context.Context) (err error)
	ListenAndServe(ctx context.Context, handler Handler) (err error) // Start server with block
	ListenAndServeTLS(ctx context.Context, addr, certFile, keyFile string, handler Handler) (err error)

	Start(ctx context.Context) (err error) // Start server without block
	Stop() (err error)

	Restart(ctx context.Context) (err error)
	HotReload(ctx context.Context) (err error)
	Shutdown() (err error) // = Stop
	Close()                // = Stop

	WithOnShutdown(cb OnShutdown) Server // set OnShutdown handler
}

type ServerOpt func(s *serverWrap)

type serverWrap struct {
	network string
	address string
	// host       string // host[:port]
	// port       int
	lc         *net.ListenConfig
	tlsConfig  *tls.Config
	bufferSize int
	quiet      bool

	onProcessData            OnTcpServerProcessData
	onCorruptData            OnTcpServerCorruptData
	onCreateReadWriter       OnTcpServerCreateReadWriter
	onConnectedWithClient    []OnTcpServerConnectedWithClient
	onDisconnectedWithClient []OnTcpServerDisconnectedWithClient
	onListening              []OnTcpServerListening
	onShutdown               OnShutdown
	onHotReload              OnHotReload
	onRestart                OnRestart
	onNewResponse            OnNewResponse

	protocolInterceptor api.ServerInterceptor

	loop          func(ctx context.Context) (err error)
	closeListener func() (err error)
	// l             net.Listener
	// pConn         net.PacketConn
	// udpConn     *net.UDPConn
	// lConn       *net.UnixConn
	handler     Handler
	connections map[*connS]bool
	exited      int32

	baseS
}

// Handler is implemented by any value that implements ServeDNS.
type Handler interface {
	Serve(ctx context.Context, w api.Response, r api.Request) (processed bool, err error)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as DNS handlers.  If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler object that calls f.
type HandlerFunc func(ctx context.Context, w api.Response, r api.Request) (processed bool, err error)

// Serve calls f(w, r).
func (f HandlerFunc) Serve(ctx context.Context, w api.Response, r api.Request) (processed bool, err error) {
	return f(ctx, w, r)
}

type OnNewResponse interface {
	New() api.Response
}

type Runnable interface {
	Run() // a go routine here
}

type CacheWriteable interface {
	// WrChannel is only available without user Handler specified.
	// The default connS.serve will received the data from WrChannel and write
	// to internal connection rawly.
	WrChannel() chan<- []byte
}

type OnTcpServerCreateReadWriter func(ss Server, conn api.Response, tsConnected time.Time) (in io.Reader, out io.Writer)
type OnTcpServerConnectedWithClient func(w api.Response, ss Server)
type OnTcpServerDisconnectedWithClient func(w api.Response, r api.Request, ss Server)
type OnTcpServerProcessData func(data []byte, w api.Response, r api.Request) (nn int, err error)
type OnTcpServerCorruptData func(data []byte, w api.Response, r api.Request) (ate int)
type OnTcpServerListening func(ss Server, l net.Listener) // for udp or unixgram, l is nil; for tcp and unix, it points to the listener
type OnShutdown func(errReason error, ss Server)
type OnHotReload func(ctx context.Context, ss Server) (err error)
type OnRestart func(ctx context.Context, ss Server) (err error)

// DataProcessor handles data and extract one or more data diagrams.
//
// If must necessary, reading more bytes from r and writing something to w.
type DataProcessor interface {
	// Process implements Processor interface to announce that i will process the incoming data in Read().
	Process(data []byte, w api.Response, r api.Request) (nn int, err error)
}

// CorruptDataFinder try finding bounds of next good diagram and returns ate bytes.
//
// If ate ok, buf[ate:] should be a good diagram.
// Or returns 0 to declare a wrong state.
type CorruptDataFinder interface {
	OnCorruptData(data []byte, w api.Response, r api.Request) (ate int)
}

func WithServerQuiet(b bool) ServerOpt {
	return func(s *serverWrap) {
		s.quiet = b
	}
}

func WithServerLogger(l Logger) ServerOpt {
	return func(s *serverWrap) {
		s.baseS.logger = l
	}
}

func WithServerLoggerHandler(h slog.Handler) ServerOpt {
	return func(s *serverWrap) {
		s.baseS.setLoggerHandler(h)
	}
}

// WithServerMaxMessageLength assumes a message/diagram have the given max-length.
//
// Specifying a proper buffer size is useful for parsing incoming data buffer efficiently.
//
// Too large buffer size takes wasted spaces and too small causes a large
// message cannot be split from incoming data stream.
//
// Default buffer size is 4096.
//
// serverWrap allocates 2 x bufferSize bytes to internal buffer, and
// reads 1 x bufferSize bytes into it. After one message extracted from
// internal buffer, the rest bytes are moved to beginning of buffer,
// and next reading position following it.
//
// The double bufferSize allows the above algorithm always works properly.
func WithServerMaxMessageLength(l int) ServerOpt {
	return func(s *serverWrap) {
		s.bufferSize = l
	}
}

func WithServerHandler(h Handler) ServerOpt {
	return func(s *serverWrap) {
		s.handler = h
	}
}

func WithServerHandlerFunc(h HandlerFunc) ServerOpt {
	return func(s *serverWrap) {
		s.handler = h
	}
}

// WithServerOnNewResponse sets a user-defined Response maker.
//
// The default ones does make new connS and run its looper. You may
// replace it with yours.
func WithServerOnNewResponse(nr OnNewResponse) ServerOpt {
	return func(s *serverWrap) {
		s.onNewResponse = nr
	}
}

// WithServerListenConfig gives a user-defined listening config structure
func WithServerListenConfig(c *net.ListenConfig) ServerOpt {
	return func(s *serverWrap) {
		s.lc = c
	}
}

// WithServerTLSConfig enables a tls link
func WithServerTLSConfig(c *tls.Config) ServerOpt {
	return func(s *serverWrap) {
		s.tlsConfig = c
	}
}

// WithNetwork sets network protocol: tcp, tcp4, tcp6, unix and unixpacket
func WithNetwork(network string) ServerOpt {
	return func(s *serverWrap) {
		s.network = network
	}
}

func WithServerOnCreateReadWriter(cb OnTcpServerCreateReadWriter) ServerOpt {
	return func(s *serverWrap) {
		s.onCreateReadWriter = cb
	}
}

func WithServerOnClientConnected(cb ...OnTcpServerConnectedWithClient) ServerOpt {
	return func(s *serverWrap) {
		s.onConnectedWithClient = append(s.onConnectedWithClient, cb...)
	}
}

func WithServerOnClientDisconnected(cb ...OnTcpServerDisconnectedWithClient) ServerOpt {
	return func(s *serverWrap) {
		s.onDisconnectedWithClient = append(s.onDisconnectedWithClient, cb...)
	}
}

func WithServerOnProcessData(cb OnTcpServerProcessData) ServerOpt {
	return func(s *serverWrap) {
		s.onProcessData = cb
	}
}

func WithServerOnCorruptData(cb OnTcpServerCorruptData) ServerOpt {
	return func(s *serverWrap) {
		s.onCorruptData = cb
	}
}

// WithServerOnListening sets callback to OnTcpServerListening so that
// you can do something while the server started ready.
func WithServerOnListening(cb ...OnTcpServerListening) ServerOpt {
	return func(s *serverWrap) {
		s.onListening = append(s.onListening, cb...)
	}
}

func WithServerOnShutdown(cb OnShutdown) ServerOpt {
	return func(s *serverWrap) {
		s.onShutdown = cb
	}
}

func (s *serverWrap) makeListener() (l net.Listener, err error) {
	l, err = net.Listen(s.network, s.address)
	if err == nil && s.tlsConfig != nil {
		l = tls.NewListener(l, s.tlsConfig)
	}
	if err == nil {
		s.closeListener = l.Close
		if s.network == "unix" || s.network == "unixpacket" {
			s.addCloseFunc(func() { _ = os.Remove(s.address) })
		}
		s.loop = func(ctx context.Context) (err error) {
			// timer := time.NewTicker(10 * time.Second)
			// defer func() {
			// 	timer.Stop()
			// 	s.debug("[tcp][server] runLoop goroutine exited.")
			// }()

			if !s.quiet {
				s.Info("Server starts listening", "at", l.Addr())
			}

			for {
				var conn net.Conn
				conn, err = l.Accept()
				if err != nil {
					if dc, db, _ := s.handleListenError(err); dc {
						continue // os.Exit(1)
					} else if db {
						break
					}
				}

				s.Debug("[serverWrap] new incoming connection", "remote", conn.RemoteAddr(), "local", conn.LocalAddr())
				if s.onNewResponse == nil {
					go newConn(s, conn).run(ctx)
				} else {
					w := s.onNewResponse.New()
					if r, ok := w.(Runnable); ok {
						go r.Run()
					} else {
						// nothing to do, we assume the OnNewResponse handled New()
						// which have already created a Response writer and run the
						// necessary looper.
					}
				}
			}
			s.Debug("[serverWrap] server's listener loop ended.")
			return
		}
	}
	return
}

func (s *serverWrap) makePacketListener() (conn net.PacketConn, err error) {
	conn, err = net.ListenPacket(s.network, s.address)
	// s.addCloser(conn)
	s.closeListener = conn.Close
	s.loop = func(ctx context.Context) (err error) {
		buf := make([]byte, s.bufferSize*2)
		s.tryInvokeOnListening(nil)
		for {
			var n int
			var ra net.Addr
			n, ra, err = conn.ReadFrom(buf)
			if err != nil {
				break
			}
			str := string(buf[:n])
			str = strings.TrimSpace(str)
			s.Debug("received packet", "remote.addr", ra.String(), "data", str)
			// if i < lim {
			// 	i++
			// 	go send(fmt.Sprintf("%d:%d(%s)", pid, i, s))
			// }
		}
		return
	}
	return
}

func (s *serverWrap) makeIpListener() (ipConn *net.IPConn, err error) {
	// if s.ipConn == nil {

	var ipAddr *net.IPAddr
	ipAddr, err = net.ResolveIPAddr(s.network, s.address) // "unix", "/tmp/echo.sock"
	if err == nil {
		ipConn, err = net.ListenIP(s.network, ipAddr)
		s.closeListener = ipConn.Close
		return
	}
	return
}

func (s *serverWrap) Close() {
	_ = s.Stop()
	s.baseS.Close()
}

func (s *serverWrap) Shutdown() (err error) {
	err = s.Stop()
	return
}

func (s *serverWrap) Stop() (err error) {
	if atomic.CompareAndSwapInt32(&s.exited, 0, 1) {
		defer s.tryInvokeOnShutdown(err)

		if s.closeListener != nil {
			s.Debug("[serverWrap] close listener")
			if err = s.closeListener(); err != nil {
				return
			}
			s.closeListener = nil
		}

		// for c := range s.connections {
		// 	if c != nil {
		// 		c.Close()
		// 	}
		// }
		s.connections = nil // NOTE that baseS.closers manage all connections
	}
	return
}

func (s *serverWrap) WithOnShutdown(cb OnShutdown) Server {
	s.onShutdown = cb
	return s
}

func (s *serverWrap) Listen(ctx context.Context) (err error) {
	//
	// Unix domain socket server and client There are three types of unix domain socket.
	//
	// "unix" corresponds to SOCK_STREAM
	// "unixdomain" corresponds to SOCK_DGRAM
	// "unixpacket" corresponds to SOCK_SEQPACKET
	//
	// When you write a “unix” or “unixpacket” server, use ListenUnix().
	//

	switch s.network {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		var l net.Listener
		if l, err = s.makeListener(); err != nil {
			s.handleError(err, "[serverWrap] cannot make tcp listener", "addr", s.address)
			return
		}
		s.tryInvokeOnListening(l)

	case "udp", "udp4", "udp6", "unixgram":
		var conn net.PacketConn
		if conn, err = s.makePacketListener(); err != nil {
			s.handleError(err, "[serverWrap] cannot make udp listener", "addr", s.address)
			return
		}
		_ = conn

	// case "unix", "unixpacket":
	// 	if err = s.makeUnixListener(); err != nil {
	// 		addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	// 		s.handleError(err, "[serverWrap] cannot make unix listener", "addr", addr)
	// 		return
	// 	}

	// case "unixgram":
	// 	err = errorsv3.Unimplemented
	// 	return

	case "ip", "ip4", "ip6":
		err = errUnimplemented // errorsv3.Unimplemented
		return
	}

	return
}

func (s *serverWrap) Start(ctx context.Context) (err error) {
	if err = s.Listen(ctx); err != nil {
		return
	}
	// go s.serveLoop(ctx, s.l)
	go s.loop(ctx)
	return
}

func (s *serverWrap) Serve(ctx context.Context) (err error) {
	// go s.serveLoop(ctx, s.l)
	go s.loop(ctx)
	return s.enterLoop(ctx)
}

func (s *serverWrap) ListenAndServe(ctx context.Context, handler Handler) (err error) {
	if err = s.Listen(ctx); err != nil {
		return
	}
	return s.Serve(ctx)
}

func (s *serverWrap) ListenAndServe1(ctx context.Context, addr string, handler Handler) (err error) {
	if err = s.Listen(ctx); err != nil {
		return
	}
	if handler != nil {
		s.handler = handler
	}
	return s.Serve(ctx)
}

func (s *serverWrap) ListenAndServeTLS(ctx context.Context, addr, certFile, keyFile string, handler Handler) (err error) {
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	if s.tlsConfig == nil {
		s.tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	} else {
		s.tlsConfig.Certificates = []tls.Certificate{cert}
	}

	s.network = "tcp-tls"
	if addr != "" {
		s.address = addr
	}

	if handler != nil {
		s.handler = handler
	}
	return
}

func (s *serverWrap) enterLoop(ctx context.Context) (err error) {
	defer s.tryInvokeOnShutdown(err)
	for {
		// block at ctx.Done and without others trying HERE!
		select {
		case <-ctx.Done():
			s.Debug("[serverWrap] server's loop ended.")
			return
		}
	}
}

func (s *serverWrap) Restart(ctx context.Context) (err error) {
	s.Close()
	if atomic.CompareAndSwapInt32(&s.exited, 1, 0) {
		if s.onRestart != nil {
			if err = s.onRestart(ctx, s); err != nil {
				return
			}
		}
		err = s.ListenAndServe(ctx, nil)
	}
	return
}

func (s *serverWrap) HotReload(ctx context.Context) (err error) {
	// reload configs and apply them
	if s.onHotReload != nil {
		err = s.onHotReload(ctx, s)
	}
	return
}

func (s *serverWrap) IsExited() bool { return atomic.LoadInt32(&s.exited) == 1 }

func (s *serverWrap) tryInvokeOnClientConnected(w api.Response) {
	s.Verbose("[serverWrap] invoke onConnected", "cb", s.onListening)
	for _, cb := range s.onConnectedWithClient {
		if cb != nil {
			cb(w, s)
		}
	}
}

func (s *serverWrap) tryInvokeOnClientDisconnected(w api.Response, r api.Request) {
	s.Verbose("[serverWrap] invoke onDisconnected", "cb", s.onListening)
	for _, cb := range s.onDisconnectedWithClient {
		if cb != nil {
			cb(w, r, s)
		}
	}
}

func (s *serverWrap) tryInvokeOnListening(l net.Listener) {
	s.Verbose("[serverWrap] invoke onListening", "cb", s.onListening)
	for _, cb := range s.onListening {
		if cb != nil {
			cb(s, l)
		}
	}
}

func (s *serverWrap) tryInvokeOnShutdown(err error) {
	if s.onShutdown != nil {
		s.Verbose("[serverWrap] loop ended, call onShutdown...")
		s.onShutdown(err, s)
		s.onShutdown = nil
	}
}

func (s *serverWrap) handleListenError(errGot error) (doContinue, doBreak bool, err error) {
	if s.IsExited() {
		doBreak = true
		return
	}

	var neterr net.Error
	if errors.As(errGot, &neterr) && (neterr.Timeout()) {
		s.Warn("[serverWrap] network error (temporary, or timeout), sleep 5ms and retry...", "err", neterr)
		time.Sleep(5 * time.Millisecond)
		doContinue, err = true, neterr
	} else {
		s.Error("error accepting: ", "err", errGot)
		time.Sleep(5 * time.Millisecond)
		doContinue, err = true, errGot // os.Exit(1)
	}
	return
}

// Client finds and returns a client's connection object
func (s *serverWrap) Client(addr string) (conn *connS) {
	for c := range s.connections {
		if c.conn.RemoteAddr().String() == addr {
			return c
		}
	}
	return
}

func (s *serverWrap) closeClient(conn *connS) {
	if _, ok := s.connections[conn]; ok {
		delete(s.connections, conn)
		conn.Close()
	}
}

func (s *serverWrap) closeClientByAddr(addr string) {
	s.closeClient(s.Client(addr))
}

//

//

//

func newConn(s *serverWrap, conn net.Conn) *connS {
	c := &connS{
		serverWrap:   s,
		conn:         conn,
		tmStart:      time.Now().UTC(),
		writeTimeout: 5 * time.Second,
		chWriteSize:  16,
		wl:           &sync.Mutex{},
	}
	c.chWrite = make(chan []byte, c.chWriteSize)
	s.connections[c] = true
	return c
}

type connS struct {
	*serverWrap
	conn         net.Conn
	tmStart      time.Time
	tmStop       time.Time
	closed       int32
	writeTimeout time.Duration
	chWrite      chan []byte
	chWriteSize  int
	wl           sync.Locker // specially for RawWrite
}

func (s *connS) WrChannel() chan<- []byte {
	return s.chWrite
}

func (s *connS) RemoteAddr() net.Addr {
	if s.conn == nil {
		return nil
	}
	return s.conn.RemoteAddr()
}

func (s *connS) RemoteAddrString() string {
	if s.conn == nil {
		return "(not-connected)"
	}
	return s.conn.RemoteAddr().String()
}

func (s *connS) LocalAddr() net.Addr {
	if s.conn == nil {
		return nil
	}
	return s.conn.LocalAddr()
}

func (s *connS) Closed() bool    { return atomic.LoadInt32(&s.closed) != 0 }
func (s *connS) NotClosed() bool { return atomic.LoadInt32(&s.closed) == 0 }
func (s *connS) Connected() bool { return s.conn != nil }
func (s *connS) Close()          { _ = s.SafeClose() }
func (s *connS) SafeClose() (err error) {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		if s.conn != nil {
			if err = s.conn.Close(); err != nil {
				s.handleError(err, "close connection failed", "client.addr", s.conn.RemoteAddr())
			}
			s.conn = nil
		}
	}
	return
}

func (s *connS) GetClientID() string {
	if s.conn == nil {
		return ""
	}
	return s.conn.RemoteAddr().String()
}

func (s *connS) Write(data []byte) (n int, err error) {
	if n = len(data); n > 0 {
		s.chWrite <- data
		s.Verbose("[connS] Write() cached one message.")
	}
	return
}

func (s *connS) RawWrite(ctx context.Context, data []byte) (n int, err error) {
	if n = len(data); n > 0 {
		n, err = s.rawWriteNow(data, s.writeTimeout)
	}
	return
}

func (s *connS) RawWriteTimeout(data []byte, deadline ...time.Duration) (n int, err error) {
	if n = len(data); n > 0 {
		durDeadline := s.writeTimeout
		for _, d := range deadline {
			durDeadline = d
		}
		n, err = s.rawWriteNow(data, durDeadline)
	}
	return
}

func (s *connS) rawWriteNow(data []byte, deadline time.Duration) (n int, err error) {
	err = s.conn.SetWriteDeadline(time.Now().Add(deadline))

	if err == nil {
		s.wl.Lock()
		defer s.wl.Unlock()
		n, err = s.conn.Write(data)
	}
	return
}

func (s *connS) Read(p []byte) (n int, err error) {
	if s.NotClosed() {
		n, err = s.conn.Read(p)
		// err = errorsv3.MethodNotAllowed // read directly is not allowed
	}
	return
}

func (s *connS) defaultProcessData(data []byte, w api.Response, r api.Request) (nn int, err error) {
	s.Debug("[connS] RECV:", "data", string(data), "client.addr", w.RemoteAddr())
	nn = len(data)
	return
}

func (s *connS) run(ctx context.Context) {
	s.tryInvokeOnClientConnected(s)
	// reader, writer := s.TryInvokeOnCreateReadWriter(s.conn, s.tmStart)
	defer s.tryInvokeOnClientDisconnected(s, s)

	if z, hasProcess := s.handler.(DataProcessor); hasProcess {
		s.onProcessData = z.Process
	}
	if s.onProcessData == nil {
		s.onProcessData = s.defaultProcessData
	}

	if z, hasCD := s.handler.(CorruptDataFinder); hasCD {
		s.onCorruptData = z.OnCorruptData
	}
	if s.onCorruptData == nil {
		s.onCorruptData = func(data []byte, w api.Response, r api.Request) (ate int) { return len(data) }
	}

	if s.handler != nil {
		if processed, err := s.handler.Serve(ctx, s, s); err != nil {
			s.Error("HandlerFunc processed failed", "err", err)
			return
		} else if processed {
			return
		}
	}

	// fallback to default serve routine
	s.serve(ctx, s, s)
}

func (s *connS) serve(ctx context.Context, w api.Response, r api.Request) {
	defer s.Close()
	s.Verbose("[connS] looper - entering...")
	go s.readBump(ctx, w, r)
writeBump:
	for {
		select {
		case <-ctx.Done():
			s.Debug("[connS] looper/writeBump ended.")
			break writeBump

		case data := <-s.chWrite:
			s.Verbose("[connS] rawWriteNow wake up")
			if len(data) == 0 {
				continue
			}
			if pi := s.protocolInterceptor; pi != nil {
				if processed, err := pi.OnWriting(ctx, s, data); processed {
					continue
				} else if err != nil {
					s.Error("[connS] Write failed", "err", err)
					break writeBump
				}
			}
			if _, err := s.rawWriteNow(data, s.writeTimeout); err != nil {
				s.handleError(err, "[connS] Write failed")
				break writeBump
			}
		}
	}

	s.tmStop = time.Now().UTC()
}

func (s *connS) readBump(ctx context.Context, w api.Response, r api.Request) {
	var nRead, pos int
	buf := make([]byte, s.bufferSize*2)
	cidHolder, ok := r.(interface{ GetClientID() string })
	if !ok || cidHolder == nil {
		cidHolder = s
	}

workingLoop:
	for {
		s.Verbose("[connS] read once", "pos", pos)
		n, err := r.Read(buf[pos : pos+s.bufferSize])
		if err != nil {
			s.handleReadError(n, err, buf, pos, w, r)
			break workingLoop
		} else if n == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		select {
		case <-ctx.Done():
			s.Debug("[connS] lopper/readBump ended.")
			break workingLoop
		default:
		}

		nEnd := pos + n
		nRead, err = s.onProcessData(buf[:nEnd], w, r)
		// s.Verbose("[connS] onProcessData processed", "nRead", nRead, "nEnd", nEnd, "err", err)

		if nRead <= 0 {
			// bad package found, skip the pieces and try to recover
			s.Warn("[connS] data block decode failed, skipped.", "client.addr", w.RemoteAddr(), "client.id", cidHolder.GetClientID(), "data", buf[:nEnd], "err", err)
			pos = s.onCorruptData(buf[:nEnd], w, r)
			if pos == nEnd {
				pos = 0
			} else {
				copy(buf, buf[pos:nEnd])
				pos = nEnd - pos
			}
			continue
		}

		if err != nil {
			s.handleError(err, "[connS] onProcessData(buf, wr) failed.", "client.addr", w.RemoteAddr(), "client.id", cidHolder.GetClientID(), "nRead", nRead)
			break workingLoop
		}

		pos = nRead
		if pos < nEnd {
			copy(buf, buf[pos:nEnd]) // move the rest bytes to the beginning of buffer
		} // else pos == nEnd
		pos = nEnd - pos // and set the ending position
	}
}

func (s *connS) handleReadError(n int, err error, buf []byte, pos int, w api.Response, r api.Request) {
	cidHolder, ok := r.(interface{ GetClientID() string })
	if !ok || cidHolder == nil {
		cidHolder = s
	}

	if errors.Is(err, io.EOF) {
		s.Debug("[connS] ♦︎ read i/o eof found. closing connection...", "client.addr", w.RemoteAddr(), "client.id", cidHolder.GetClientID())
		return
	}
	if n > 0 {
		_, _ = s.onProcessData(buf[:pos+n], w, r)
	}
	if strings.Contains(err.Error(), "use of closed network connection") {
		s.Trace("[connS] ♦︎ closed by others.", "client.addr", w.RemoteAddrString())
	} else {
		s.handleError(err, "[connS] ♦︎︎ reader.read(buf) failed. closing", "client.addr", w.RemoteAddr(), "client.id", cidHolder.GetClientID())
	}
}

func (s *connS) handleError(err error, reason string, args ...any) {
	s.baseS.handleError(err, reason, args...)
	if s.NotClosed() {
		err = checkConn(s.conn) // try inspecting raw error
		if err != nil {
			s.Error("ERROR", "err-after-check-conn", err)
		}
	}
}
