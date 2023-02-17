package udp

import (
	"context"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/log"
	"gopkg.in/hedzr/go-ringbuf.v1/fast"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
)

type Obj struct {
	log.Logger
	protocol.InterceptorHolder

	WriteTimeout time.Duration

	rb fast.RingBuffer

	addr           *net.UDPAddr
	udpconn        *net.UDPConn
	baseConn       base.Conn
	maxBufferSize  int
	rdCh           chan *base.UdpPacket
	wrCh           chan *base.UdpPacket
	debugMode      bool
	wg             sync.WaitGroup
	listenerNumber int
	closed         int32
}

func (s *Obj) IsConnected() bool {
	return s.baseConn != nil
}

func (s *Obj) AsBaseConn() base.Conn {
	return s.baseConn
}

// Join will wait for all internal pumps stopped and close done chan bool
func (s *Obj) Join(ctx context.Context, done chan bool) {
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
			if s.udpconn != nil {
				err = s.udpconn.Close()
				s.udpconn = nil
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
func (s *Obj) Connect(baseCtx context.Context, config *base.Config) (err error) {
	var sip, sport string
	var port int
	sip, sport, err = net.SplitHostPort(config.Addr)
	if err == nil {

		var ip net.IP
		if sip == "" {
			if config.Network == "udp6" {
				ip = net.IPv6zero
			} else {
				ip = net.IPv4zero
			}
		} else {
			ip = net.ParseIP(sip)
			if ip == nil {
				var udpAddr *net.UDPAddr
				udpAddr, err = net.ResolveUDPAddr(config.Network, sip)
				if err != nil {
					var ipAddr *net.IPAddr
					ipAddr, err = net.ResolveIPAddr("ip", sip)
					if err != nil {
						return
					}
					ip = ipAddr.IP
				} else {
					ip = udpAddr.IP
				}
			}
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

			// ctx, cancel := context.WithDeadline(baseCtx, time.Now().Add(10*time.Second))
			// defer cancel()

			s.udpconn, err = net.DialUDP(config.Network, srcAddr, s.addr)
			if err == nil {
				// s.connected = true
				s.baseConn = &udpConnWrapper{s, s.udpconn, s.Logger}
				s.Debugf("Connecting OK: %v / %v", config.Addr, config.Uri)
				// err = s.conn.SetWriteBuffer(8192)
				if s.ProtocolInterceptor() != nil {
					s.ProtocolInterceptor().OnConnected(baseCtx, s.baseConn)
				}
			}
		}
	}
	return
}

// Create a server listener via net.ListenUDP()
func (s *Obj) Create(baseCtx context.Context, network string, config *base.Config) (err error) {
	// var udpConn *net.UDPConn
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
				s.udpconn, err = net.ListenMulticastUDP(network, en4, s.addr)
			} else {
				s.udpconn, err = net.ListenUDP(network, s.addr)
				if err == nil {
					err = s.udpconn.SetReadBuffer(1048576)
				}
			}
			if err == nil && s.udpconn != nil && s.ProtocolInterceptor() != nil {
				s.baseConn = &udpConnWrapper{s, s.udpconn, s.Logger}
				s.Tracef("Created OK: %v", config.Addr)
				s.ProtocolInterceptor().OnListened(baseCtx, config.Addr)
			}
		}
	}
	return
}

func (s *Obj) ClientServe(baseCtx context.Context) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	if s.listenerNumber < 1 {
		s.listenerNumber = runtime.NumCPU()
	}
	for i, max := 0, s.listenerNumber; i < max; i++ {
		ctx1 := context.WithValue(ctx, "tid", i)
		go s.clientListener(ctx1)
	}

	go s.readPump(ctx)
	s.writePump(ctx)

	// writePump will be end while wrCh closed (from Close())
	// and defered cancel() will stop the readPump loop
	return
}

func (s *Obj) Serve(baseCtx context.Context) (err error) {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	if s.listenerNumber < 1 {
		s.listenerNumber = runtime.NumCPU()
	}
	for i, max := 0, s.listenerNumber; i < max; i++ {
		ctx1 := context.WithValue(ctx, "tid", i)
		go s.listener(ctx1)
	}

	go s.readPump(ctx)
	s.writePump(ctx)

	// writePump will be end while wrCh closed (from Close())
	// and defered cancel() will stop the readPump loop
	return
}

func (s *Obj) RemoteAddr() net.Addr {
	return s.baseConn.RemoteAddr()
}

// ReadThrough read binary data to a connected *UDPConn (via udp.Obj.Create())
func (s *Obj) ReadThrough(data []byte) (n int, err error) {
	return s.udpconn.Read(data)
}

func (s *Obj) RawWrite(ctx context.Context, data []byte) (n int, err error) {
	deadline := time.Now().Add(s.WriteTimeout)
	err = s.udpconn.SetWriteDeadline(deadline)
	if err != nil {
		return
	}
	return s.udpconn.Write(data)
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
		s.Tracef("[udp.doWrite] writing: %v / %v", string(packet.Data), packet)
		deadline := time.Now().Add(s.WriteTimeout)
		err = s.udpconn.SetWriteDeadline(deadline)
		if err != nil {
			return
		}
		_, err = s.udpconn.Write(packet.Data)
		if err != nil {
			s.Errorf("[udp.doWrite] written error: %v", err)
		}
		return
	}

	// s.conn.SetWriteBuffer(33)
	s.Tracef("[udp.doWrite] WriteToUDP: %v / %v", string(packet.Data), packet)
	_, err = s.udpconn.WriteToUDP(packet.Data, packet.RemoteAddr)
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
			if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
				s.InterceptorHolder.ProtocolInterceptor().OnError(ctx, s.baseConn, err)
			}
		}
		s.wg.Done()
	}()

	// // raise the connected event now
	// if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
	//	s.InterceptorHolder.ProtocolInterceptor().OnConnected(ctx, s.baseConn)
	// }

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
				if s.InterceptorHolder != nil && s.InterceptorHolder.ProtocolInterceptor() != nil {
					s.InterceptorHolder.ProtocolInterceptor().OnError(ctx, s.baseConn, err)
				}
			}
		}
	}
}

func (s *Obj) listener(baseCtx context.Context) {
	buffer := make([]byte, s.maxBufferSize)
	retry, noDebug, n, remoteAddr, err := 0, s.debugMode, 0, new(net.UDPAddr), error(nil)

	s.wg.Add(1)

	defer func() {
		if err == nil {
			s.Debugf("    .. [udp.listener] #%-5d listener end", baseCtx.Value("tid"))
		} else {
			s.Errorf("    .. [udp.listener] #%-5d listener failed - %v", baseCtx.Value("tid"), err)
		}
		s.wg.Done()
	}()

	for err == nil {
		n, remoteAddr, err = s.udpconn.ReadFromUDP(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// s.Tracef("    .. [udp.listener] #%-5d client conn closed, remote is %q", baseCtx.Value("tid"), s.conn.RemoteAddr())
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
		// because you've only made one buffer per listener().
		//
		// fmt.Println("from", remoteAddr, "-", buffer[:n])
		sd := make([]byte, n)
		copy(sd, buffer[:n])
		s.Tracef("[udp.listener] #%-5d : %v -> %v %q", baseCtx.Value("tid"), remoteAddr, sd, sd)
	retryPut:
		err = s.rb.Enqueue(base.NewUdpPacket(remoteAddr, sd))
		if err != nil {
			if err == fast.ErrQueueFull {
				// block till queue not full
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if retry > 1000 {
					if !noDebug && retry > 1000 {
						s.Warnf("[udp.listener] #%-5d (retry %v). quantity = %v. FULL! block till queue not full.", baseCtx.Value("tid"), retry, s.rb.Quantity())
					}
					break
				}
				goto retryPut
			}
			s.Errorf("[udp.listener] #-5d err: %+v.", baseCtx.Value("tid"), err)
			continue
		}
		s.Tracef("[udp.listener] #%-5d : %v -> %v %q | enqueued", baseCtx.Value("tid"), remoteAddr, sd, sd)
	}
}

func (s *Obj) clientListener(baseCtx context.Context) {
	buffer := make([]byte, s.maxBufferSize)
	retry, noDebug, n, remoteAddr, err := 0, s.debugMode, 0, new(net.UDPAddr), error(nil)

	remoteAddr = s.udpconn.RemoteAddr().(*net.UDPAddr)

	s.wg.Add(1)

	defer func() {
		if err == nil {
			s.Debugf("    .. [udp.listener.client] #%-5d listener end", baseCtx.Value("tid"))
		} else {
			s.Errorf("    .. [udp.listener.client] #%-5d listener failed - %v", baseCtx.Value("tid"), err)
		}
		s.wg.Done()
	}()

	for err == nil {
		n, err = s.udpconn.Read(buffer)
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				// s.Tracef("    .. [udp.listener.client] #%-5d client conn closed, remote is %q", baseCtx.Value("tid"), s.conn.RemoteAddr())
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
		// because you've only made one buffer per listener().
		//
		// fmt.Println("from", remoteAddr, "-", buffer[:n])
		sd := make([]byte, n)
		copy(sd, buffer[:n])
		s.Tracef("[udp.listener.client] #%-5d : %v -> % x %q", baseCtx.Value("tid"), remoteAddr, sd, sd)
	retryPut:
		err = s.rb.Enqueue(base.NewUdpPacket(remoteAddr, sd))
		if err != nil {
			if err == fast.ErrQueueFull {
				// block till queue not full
				time.Sleep(time.Duration(retry) * time.Microsecond)
				retry++
				if retry > 1000 {
					if !noDebug && retry > 1000 {
						s.Warnf("[udp.listener.client] #%-5d (retry %v). quantity = %v. FULL! block till queue not full.", baseCtx.Value("tid"), retry, s.rb.Quantity())
					}
					break
				}
				goto retryPut
			}
			s.Errorf("[udp.listener.client] #-5d err: %+v.", baseCtx.Value("tid"), err)
			continue
		}
		s.Tracef("[udp.listener.client] #%-5d : %v -> %v %q | enqueued", baseCtx.Value("tid"), remoteAddr, sd, sd)
	}
}

const (
	DefaultPacketSize      = 4096
	DefaultPacketQueueSize = 1024
)
