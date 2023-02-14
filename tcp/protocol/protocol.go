package protocol

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
)

type ClientInterceptor interface {
	OnConnected(ctx context.Context, c base.Conn)
	OnClosing(c base.Conn, reason int)
	OnClosed(c base.Conn, reason int)

	OnError(ctx context.Context, c base.Conn, err error)

	// OnReading handles reading event for tcp mode
	OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	// OnWriting handles writing event for tcp mode
	// You may override the internal writing action with processed = true and 
	// write data yourself. for instance:
	//     processed = true
	//     c.RawWrite(data)
	OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	
	// OnUDPReading is special hook if in udp mode
	OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
	// OnUDPWriting is special hook if in udp mode
	OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
}

type Interceptor interface {
	OnListened(baseCtx context.Context, addr string)
	OnServerReady(ctx context.Context, c log.Logger)
	OnServerClosed(server log.Logger)

	ClientInterceptor
}

type ClientInterceptorHolder interface {
	ProtocolInterceptor() ClientInterceptor
}

type InterceptorHolder interface {
	ProtocolInterceptor() Interceptor
}
