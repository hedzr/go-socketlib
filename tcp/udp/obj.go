package udp

import (
	"context"
	"github.com/hedzr/go-ringbuf/fast"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Obj struct {
	log.Logger
	protocol.InterceptorHolder
	conn          *net.UDPConn
	addr          *net.UDPAddr
	baseConn      base.Conn
	maxBufferSize int
	rb            fast.RingBuffer
	debugMode     bool
	connected     bool
	rdCh          chan *base.UdpPacket
	wrCh          chan *base.UdpPacket
	WriteTimeout  time.Duration
	wg            sync.WaitGroup
	closed        int32
}

func NewUdpObj(so protocol.InterceptorHolder, conn *net.UDPConn, addr *net.UDPAddr) *Obj {
	if x := fast.New(DefaultPacketQueueSize,
		fast.WithDebugMode(false),
		fast.WithLogger(so.(log.Logger)),
	); x != nil {
		return &Obj{
			Logger:            so.(log.Logger),
			InterceptorHolder: so,
			conn:              conn,
			addr:              addr,
			maxBufferSize:     DefaultPacketSize,
			rb:                x,
			debugMode:         false,
			rdCh:              make(chan *base.UdpPacket, DefaultPacketQueueSize),
			wrCh:              make(chan *base.UdpPacket, DefaultPacketQueueSize),
			WriteTimeout:      10 * time.Second,
		}
	}
	return nil
}

func (s *Obj) IsConnected() bool {
	return s.connected
}

func (s *Obj) AsBaseConn() base.Conn {
	return s.baseConn
}

func (s *Obj) Join(ctx context.Context, done chan struct{}) {
	if s.baseConn != nil {
		go func() { time.Sleep(100 * time.Millisecond); _ = s.Close() }()
		s.wg.Wait()
		close(done)
	}
}

func (s *Obj) Close() (err error) {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		if s.baseConn != nil {
			if s.ProtocolInterceptor() != nil {
				s.ProtocolInterceptor().OnClosing(s.baseConn, 0)
			}
			close(s.rdCh)
			close(s.wrCh)
			if s.conn != nil {
				err = s.conn.Close()
				s.conn = nil
				if err != nil {
					s.Debugf("s.conn closed: %v", err)
				}
			}
			if s.ProtocolInterceptor() != nil {
				s.ProtocolInterceptor().OnClosed(s.baseConn, 0)
			}
			s.baseConn = nil
		}
	}
	return
}

// Connect to a server endpoint via net.DialUDP()
func (s *Obj) Connect(baseCtx context.Context, network string, config *base.Config) (err error) {
	var sip, sport string
	var port int
	sip, sport, err = net.SplitHostPort(config.Addr)
	if err == nil {

		var ip net.IP
		if sip == "" {
			if network == "udp6" {
				ip = net.IPv6zero
			} else {
				ip = net.IPv4zero
			}
		} else {
			ip = net.ParseIP(sip)
		}

		port, err = strconv.Atoi(sport)

		if err == nil {
			s.addr = &net.UDPAddr{
				IP:   ip,
				Port: port,
			}
			srcAddr := &net.UDPAddr{
				IP:   net.IPv4zero,
				Port: 0,
				Zone: "",
			}

			//ctx, cancel := context.WithDeadline(baseCtx, time.Now().Add(10*time.Second))
			//defer cancel()

			s.conn, err = net.DialUDP(network, srcAddr, s.addr)
			if err == nil {
				s.connected = true
				s.baseConn = &udpConnWrapper{s.conn, s.Logger}
				//err = s.conn.SetWriteBuffer(8192)
				if s.ProtocolInterceptor() != nil {
					s.ProtocolInterceptor().OnConnected(baseCtx, s.baseConn)
				}
			}
		}
	}
	return
}

// Create a server listener via net.ListenUDP()
func (s *Obj) Create(network string, config *base.Config) (err error) {
	//var udpConn *net.UDPConn
	var sip, sport string
	var port int
	sip, sport, err = net.SplitHostPort(config.Addr)
	if err == nil {

		var ip net.IP
		if sip == "" {
			if network == "udp6" {
				ip = net.IPv6zero
			} else {
				ip = net.IPv4zero
			}
		} else {
			ip = net.ParseIP(sip)
		}

		port, err = strconv.Atoi(sport)

		if err == nil {
			s.addr = &net.UDPAddr{
				IP:   ip,
				Port: port,
			}
			if ip.IsLinkLocalMulticast() {
				var en4 *net.Interface
				if config.Adapter != "" {
					en4, err = net.InterfaceByName(config.Adapter)
					if err != nil {
						s.Errorf("network adapter %q not found: %v", config.Adapter, err)
						return
					}
				}
				s.conn, err = net.ListenMulticastUDP(network, en4, s.addr)
			} else {
				s.conn, err = net.ListenUDP(network, s.addr)
			}
		}
	}
	return
}

func (s *Obj) Serve(baseCtx context.Context) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	for i, max := 0, runtime.NumCPU(); i < max; i++ {
		ctx1 := context.WithValue(ctx, "tid", i)
		go s.listen(ctx1)
	}

	go s.readPump(ctx)
	s.writePump(ctx)

	// writePump will be end while wrCh closed (from Close())
	// and defered cancel() will stop the readPump loop
	return
}

func (s *Obj) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

// ReadThrough read binary data to a connected *UDPConn (via udp.Obj.Create())
func (s *Obj) ReadThrough(data []byte) (n int, err error) {
	return s.conn.Read(data)
}

func (s *Obj) RawWrite(ctx context.Context, data []byte) (n int, err error) {
	deadline := time.Now().Add(s.WriteTimeout)
	err = s.conn.SetWriteDeadline(deadline)
	if err != nil {
		return
	}
	return s.conn.Write(data)
}

// WriteThrough send binary data to a connected *UDPConn (via udp.Obj.Create())
func (s *Obj) WriteThrough(data []byte) (n int, err error) {
	return s.RawWrite(context.Background(), data)
}

// Write sent udp-packet to a known peer asynchronously
func (s *Obj) Write(data *base.UdpPacket) {
	s.wrCh <- data
}

// WriteTo sent udp-packet to a known peer asynchronously.
// remoteAddr can be nil if sending for a client connected by udp.Create().
func (s *Obj) WriteTo(remoteAddr *net.UDPAddr, data []byte) {
	s.wrCh <- &base.UdpPacket{
		RemoteAddr: remoteAddr,
		Data:       data,
	}
}

func (s *Obj) doWrite(ctx context.Context, packet *base.UdpPacket) (err error) {
	if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
		var processed bool
		if processed, err = s.InterceptorHolder.ProtocolInterceptor().OnUDPWriting(ctx, s, packet); processed {
			return
		} else if err != nil {
			s.Warnf("[udp.doWrite] protocolInterceptor got error: %v", err)
			return
		}
	}

	if packet.RemoteAddr == nil {
		s.Debugf("[udp.doWrite] writing: %v / %v", string(packet.Data), packet)
		deadline := time.Now().Add(s.WriteTimeout)
		err = s.conn.SetWriteDeadline(deadline)
		if err != nil {
			return
		}
		_, err = s.conn.Write(packet.Data)
		if err != nil {
			s.Errorf("[udp.doWrite] written error: %v", err)
		}
		return
	}

	//s.conn.SetWriteBuffer(33)
	s.Debugf("[udp.doWrite] WriteToUDP: %v / %v", string(packet.Data), packet)
	_, err = s.conn.WriteToUDP(packet.Data, packet.RemoteAddr)
	if err != nil {
		s.Errorf("[udp.doWrite] WriteToUDP error: %v", err)
	}
	return
}

func (s *Obj) writePump(ctx context.Context) {
	var err error

	s.wg.Add(1)
	defer func() {
		if err == nil {
			s.Debugf("    .. writePump end.")
		} else {
			s.Errorf("    .. writePump end with error: %v", err)
		}
		s.wg.Done()
	}()

	for err == nil {
		select {
		case <-ctx.Done():
			s.Debugf("    .. writePump will be end by ctx.Done.")
			return
		case data := <-s.wrCh:
			if data == nil {
				return
			}
			if err = s.doWrite(ctx, data); err != nil {
				s.Errorf("internal write failed: %v", err)
			}
		}
	}
}

func (s *Obj) readPump(ctx context.Context) {
	var (
		err       error
		it        interface{}
		processed bool
		retry     = 0
	)

	s.wg.Add(1)

	defer func() {
		if err == nil {
			s.Debugf("    .. readPump end.")
		} else {
			s.Errorf("    .. readPump end with error: %v", err)
		}
		s.wg.Done()
	}()

	for err == nil {
		select {
		case <-ctx.Done():
			return
		default:
		}

		it, err = s.rb.Dequeue()
		if err != nil {
			if err == fast.ErrQueueEmpty {
				// block till queue not empty
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if s.debugMode && retry > 2000 {
					s.Warnf("[udp.readPump] (retry %v). quantity = %v. EMPTY! block till queue not empty.", retry, s.rb.Quantity())
				}
				err = nil
				continue
			}
			s.Errorf("[udp.readPump] failed. err: %+v.", err)
			if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
				s.InterceptorHolder.ProtocolInterceptor().OnError(ctx, nil, err)
			}
		}

		retry = 1
		if packet, ok := it.(*base.UdpPacket); ok {
			if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
				if processed, err = s.InterceptorHolder.ProtocolInterceptor().OnUDPReading(ctx, s, packet); processed {
					continue
				} else if err != nil {
					s.Warnf("[udp.readPump] protocolInterceptor got error: %v", err)
					err = nil
					continue
				}
			}

			s.Tracef("[udp.readPump] %v -> % x %q", packet.RemoteAddr, packet.Data, string(packet.Data))
			if s.IsConnected() == false {
				s.Write(packet) // echo server
			}
		}
		// t.Logf("[GET] %5d. '%v' GOT, quantity = %v.", i, it, fast.Quantity())
	}
}

func (s *Obj) listen(baseCtx context.Context) {
	buffer := make([]byte, s.maxBufferSize)
	retry, noDebug, n, remoteAddr, err := 0, s.debugMode, 0, new(net.UDPAddr), error(nil)

	s.wg.Add(1)

	defer func() {
		if err == nil {
			s.Debugf("    .. [udp.listen.routine] #%-5d listener end", baseCtx.Value("tid"))
		} else {
			s.Errorf("    .. [udp.listen.routine] #%-5d listener failed - %v", baseCtx.Value("tid"), err)
		}
		s.wg.Done()
	}()

	for err == nil {
		n, remoteAddr, err = s.conn.ReadFromUDP(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// s.Tracef("    .. [udp.listen.routine] #%-5d client conn closed, remote is %q", baseCtx.Value("tid"), s.conn.RemoteAddr())
				err = nil
				return
			}
			return
		}

		select {
		case <-baseCtx.Done():
			return
		default:
		}

		// you might copy out the contents of the packet here, to
		// `var r myapp.Request`, say, and `go handleRequest(r)` (or
		// send it down a channel) to free up the listening
		// goroutine. you do *need* to copy then, though,
		// because you've only made one buffer per listen().
		//
		// fmt.Println("from", remoteAddr, "-", buffer[:n])
		sd := make([]byte, n)
		copy(sd, buffer[:n])
		s.Debugf("[udp.listen.routine] #%-5d : %v -> %v %q", baseCtx.Value("tid"), remoteAddr, sd, sd)
	retryPut:
		err = s.rb.Enqueue(base.NewUdpPacket(remoteAddr, sd))
		if err != nil {
			if err == fast.ErrQueueFull {
				// block till queue not full
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if retry > 1000 {
					if !noDebug && retry > 1000 {
						s.Warnf("[udp.listen.routine] #%-5d (retry %v). quantity = %v. FULL! block till queue not full.", baseCtx.Value("tid"), retry, s.rb.Quantity())
					}
					break
				}
				goto retryPut
			}
			s.Errorf("[udp.listen.routine] #-5d err: %+v.", baseCtx.Value("tid"), err)
			continue
		}
		s.Tracef("[udp.listen.routine] #%-5d : %v -> %v %q | enqueued", baseCtx.Value("tid"), remoteAddr, sd, sd)
	}
}

const (
	DefaultPacketSize      = 4096
	DefaultPacketQueueSize = 1024
)
