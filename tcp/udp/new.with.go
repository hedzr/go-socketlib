package udp

import (
	"net"
	"time"

	"github.com/hedzr/log"
	"gopkg.in/hedzr/go-ringbuf.v1/fast"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
)

type Opt func(*Obj)

func New(so protocol.InterceptorHolder, opts ...Opt) (obj *Obj) {
	if x := fast.New(DefaultPacketQueueSize,
		fast.WithDebugMode(false),
		fast.WithLogger(so.(log.Logger)),
	); x != nil {
		obj = &Obj{
			Logger:            so.(log.Logger),
			InterceptorHolder: so,
			addr:              nil,
			maxBufferSize:     DefaultPacketSize,
			rb:                x,
			debugMode:         false,
			rdCh:              make(chan *base.UdpPacket, DefaultPacketQueueSize),
			wrCh:              make(chan *base.UdpPacket, DefaultPacketQueueSize),
			WriteTimeout:      10 * time.Second,
		}
	}

	for _, opt := range opts {
		if obj != nil && opt != nil {
			opt(obj)
		}
	}

	return
}

func WithListenerNumber(n int) Opt {
	return func(obj *Obj) {
		obj.listenerNumber = n
	}
}

func WithUDPConn(conn *net.UDPConn, addr *net.UDPAddr) Opt {
	return func(obj *Obj) {
		obj.udpconn, obj.addr = conn, addr
	}
}
