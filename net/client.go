package net

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/go-socketlib/net/api"
)

func NewClient(opts ...ClientOpt) *clientS {
	s := &clientS{
		dialTimeout:  15 * time.Second,
		writeTimeout: 5 * time.Second,
		bufferSize:   defaultBufferSize,
		chWriteSize:  16,
		wl:           &sync.Mutex{},
		baseS:        newBaseS(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.chWrite = make(chan []byte, s.chWriteSize)
	s.chPkg = make(chan []byte, s.chWriteSize)
	return s
}

type ClientOpt func(s *clientS)

func WithClientInterceptor(ci api.Interceptor) ClientOpt {
	return func(s *clientS) {
		s.protocolInterceptor = ci
	}
}

func WithClientLogger(l Logger) ClientOpt {
	return func(s *clientS) {
		s.baseS.logger = l
	}
}

func WithClientLoggerHandler(h slog.Handler) ClientOpt {
	return func(s *clientS) {
		s.baseS.setLoggerHandler(h)
	}
}

func WithClientSendCacheSize(size int) ClientOpt {
	return func(s *clientS) {
		s.chWriteSize = size
	}
}

type clientS struct {
	dialTimeout         time.Duration
	writeTimeout        time.Duration
	interval            time.Duration
	protocolInterceptor api.Interceptor //
	conn                net.Conn
	closed              int32
	bufferSize          int
	chWriteSize         int
	chWrite             chan []byte // cacheable writing channel
	chPkg               chan []byte // reserved for package parsing
	wl                  sync.Locker
	timesTimeout        int

	baseS
}

const maxTimesTimeout = 300

type Client interface {
	api.Addressable

	Close()
	Closed() bool
	NotClosed() bool
	Connected() bool

	Dial(network, addr string) (err error)

	api.Writeable

	io.Reader

	Run(ctx context.Context) // start a go routine to run internal worker loop

	// RunDemo(ctx context.Context) // for testing only

	// api.RawWriteable
}

func (c *clientS) RemoteAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.RemoteAddr()
}
func (c *clientS) RemoteAddrString() string {
	if c.conn == nil {
		return "(not-connected)"
	}
	return c.conn.RemoteAddr().String()
}
func (c *clientS) LocalAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.LocalAddr()
}

func (c *clientS) ProtocolInterceptor() api.Interceptor      { return c.protocolInterceptor }
func (c *clientS) SetProtocolInterceptor(pi api.Interceptor) { c.protocolInterceptor = pi }
func (c *clientS) Closed() bool                              { return atomic.LoadInt32(&c.closed) != 0 }
func (c *clientS) NotClosed() bool                           { return atomic.LoadInt32(&c.closed) == 0 }
func (c *clientS) Connected() bool                           { return c.conn != nil }

// Close makes clientS shutting down itself.
func (c *clientS) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		if c.conn != nil {
			if err := c.conn.Close(); err != nil {
				c.Error("close client failed", "err", err)
			}
			c.conn = nil
		}
	}
	c.baseS.Close()
	return
}

func (c *clientS) Dial(network, addr string) (err error) {
	c.conn, err = net.DialTimeout(network, addr, c.dialTimeout)
	return
}

func (c *clientS) Write(data []byte) (n int, err error) {
	if n = len(data); n > 0 {
		c.chWrite <- data
		c.Trace("[client] Write() cached one message.")
	}
	return
}

func (c *clientS) RawWrite(ctx context.Context, data []byte) (n int, err error) {
	if n = len(data); n > 0 {
		n, err = c.rawWriteNow(data, c.writeTimeout)
	}
	return
}

func (c *clientS) RawWriteTimeout(data []byte, deadline ...time.Duration) (n int, err error) {
	if n = len(data); n > 0 {
		durDeadline := c.writeTimeout
		for _, d := range deadline {
			durDeadline = d
		}
		n, err = c.rawWriteNow(data, durDeadline)
	}
	return
}

func (c *clientS) rawWriteNow(data []byte, deadline time.Duration) (n int, err error) {
	err = c.conn.SetWriteDeadline(time.Now().Add(deadline))

	if err == nil {
		c.wl.Lock()
		defer c.wl.Unlock()
		n, err = c.conn.Write(data)
	}
	return
}

func (c *clientS) Read(p []byte) (n int, err error) {
	// n, err = c.conn.Read(p)
	err = errMethodNotAllowed // errorsv3.MethodNotAllowed // read directly is not allowed
	return
}

func (c *clientS) Run(ctx context.Context) {
	go c.runLoop(ctx)
}

func (c *clientS) runLoop(ctx context.Context) {
	defer c.Close()
	c.Verbose("[client] looper - entering...")
	go c.readBump(ctx)

writeBump:
	for c.NotClosed() {
		select {
		case <-ctx.Done():
			c.Debug("[client] looper - ended.")
			break writeBump

		case _ = <-c.chPkg:
			// obsolete the split package

		case data := <-c.chWrite:
			c.Verbose("[client] rawWriteNow() wake up.")
			if c.NotClosed() {
				if pi, ld := c.protocolInterceptor, len(data); pi != nil && ld > 0 {
					if processed, err := pi.OnWriting(ctx, c, data); processed {
						continue
					} else if err != nil {
						c.Error("[client] Write failed", "err", err)
						break writeBump
					}
				}
				if _, err := c.rawWriteNow(data, c.writeTimeout); err != nil {
					c.Error("[client] Write failed", "err", err)
					break writeBump
				}
			}
		}
	}
}

func (c *clientS) readBump(ctx context.Context) {
	buf := make([]byte, c.bufferSize)
workingLoop:
	for c.NotClosed() {
		c.Verbose("[client] looper/readBump - reading...")
		if n, err := c.conn.Read(buf); err != nil {
			if errors.Is(err, io.EOF) {
				if n > 0 {
					c.Warn("[client]    tcp: EOF reached with some bytes", "how-many-bytes", n)
				}
				if c.Closed() {
					c.Debug("[client]     tcp: EOF reached. socket broken or closed")
					break workingLoop
				}
				err = checkConn(c.conn) // try inspecting raw error
				if err != nil && !errors.Is(err, io.EOF) {
					c.Error("[client]     tcp: checked underlying connection", "err", err)
				} else {
					c.Debug("[client]     tcp: EOF reached. cancel reading.")
				}
				time.Sleep(30 * time.Millisecond)
				break workingLoop // can't recovery from this point, exit and close socket right now
			}

			var e net.Error
			if errors.As(err, &e) && e.Timeout() {
				time.Sleep(30 * time.Millisecond)
				c.timesTimeout++
				if c.timesTimeout >= maxTimesTimeout {
					c.Warn("timeout retry over limited", "max-times-timeout", c.timesTimeout)
					break workingLoop
				}
				continue
			}

			if strings.Contains(err.Error(), "use of closed network connection") {
				c.Error("[client]  closed by others.", "server.addr", c.RemoteAddr())
			} else if strings.Contains(err.Error(), "connection reset by peer") {
				c.Error("[client]  closed by peer.", "server.addr", c.RemoteAddr())
			} else {
				c.Error("[client] Read failed", "err", err)
			}
			break workingLoop
		} else {
			select {
			case <-ctx.Done():
				c.Debug("[client] lopper/readBump ended.")
				break workingLoop
			default:
			}

			if _, err = c.tryHandleData(ctx, buf[:n]); err != nil {
				c.Error("[client] Process data failed", "err", err)
				break workingLoop
			}
		}
		c.timesTimeout = 0
	}
}

func (c *clientS) tryHandleData(ctx context.Context, data []byte) (processed bool, err error) {
	if pi, ld := c.protocolInterceptor, len(data); pi != nil && ld > 0 {
		processed, err = pi.OnReading(ctx, c, data, c.chPkg)
	} else if ld == 0 {
		time.Sleep(30 * time.Millisecond)
	} else {
		c.Trace("[client]   tryHandleData processed data once", "how-many-bytes", len(data))
	}
	return
}

func (c *clientS) RunDemo(ctx context.Context) {
	c.Run(ctx)
	c.runDemoLoop(ctx, -1)
}

func (c *clientS) runDemoLoop(ctx context.Context, count int) {
	if count < 0 {
		count = math.MaxInt
	}

	if c.interval == 0 {
		c.interval = time.Second
	}

	ticker := time.NewTicker(c.interval)
	defer func() {
		c.Debug("[client] run demo loop ended.")
		ticker.Stop()
		c.Close()
	}()

	c.Verbose("[client] run demo loop", "interval", c.interval)
	i := 0
	for i < count {
		c.Trace("[client] for", "i", i)
		select {
		case <-ctx.Done():
			i = count
		case tick := <-ticker.C:
			i++
			c.Trace("[client] tick", "tick", tick, "i", i)
			if err := c.runOnce(ctx, i); err != nil {
				c.Error("[client] send data failed", "err", err)
				i = count
			}
		}
	}
}

func (c *clientS) runOnce(ctx context.Context, count int) (err error) {
	rs := randomStringPureRange(16, 128)
	data := []byte(rs)
	c.Verbose("[client] SEND: ", "i", count, "len", len(data), "data", rs)
	_, err = c.RawWriteTimeout(data)
	if err != nil {
		c.Error("RawWriteTimeout failed", "err", err)
	}
	c.Verbose("[client] SENT: ", "i", count, "len", len(data), "data", rs)
	return
}

// randomStringPureRange generate a random string with length specified.
func randomStringPureRange(minL, maxL int) (result string) {
	source := rand.NewSource(time.Now().UnixNano())
	n := int64(float64(source.Int63())/(float64(math.MaxInt64)/float64(maxL-minL))) + int64(minL)
	b := make([]byte, int(n))
	for i := range b {
		b[i] = Alphabets[source.Int63()%int64(len(Alphabets))]
	}
	return string(b)
}

const (
	// Alphabets gets the a to z and A to Z
	Alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)
