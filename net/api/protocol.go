package api

import (
	"context"
)

type Codec interface {
	OnDecode(data []byte, ch chan<- []byte) (processed bool, err error)
	OnEncode(body []byte) (data []byte, err error)
}

type Interceptor interface {
	// OnReading handles reading event for tcp mode.
	OnReading(ctx context.Context, conn Conn, data []byte, ch chan<- []byte) (processed bool, err error)
	// OnWriting handles writing event for tcp mode.
	//
	// You may override the internal writing action with processed = true and
	// write data yourself. For instance:
	//     processed = true
	//     c.RawWrite(data)
	// By default, after OnWriting do nothing, internal loop will write data
	// to tcp connection rawly.
	OnWriting(ctx context.Context, conn Conn, data []byte) (processed bool, err error)
}

type UdpInterceptor interface {
	// OnUdpReading is special hook if in udp mode
	OnUdpReading(ctx context.Context, packet *UdpPacket) (processed bool, err error)
	// OnUdpWriting is special hook if in udp mode
	OnUdpWriting(ctx context.Context, packet *UdpPacket) (processed bool, err error)
}

type ConnAware interface {
	OnConnected(ctx context.Context, conn Conn)
	OnClosing(c Conn, reason int)
	OnClosed(c Conn, reason int)
}

type ErrorAware interface {
	OnError(ctx context.Context, conn Conn, err error)
}

type ServerInterceptor interface {
	OnListened(baseCtx context.Context, addr string)
	OnServerReady(ctx context.Context) // ctx.Value["conn"] -> api.Conn
	OnServerClosed()                   // ctx.Value["logger"] -> log/slog or logg/slog

	Interceptor
}

type InterceptorHolder interface {
	ProtocolInterceptor() Interceptor
}

type ServerInterceptorHolder interface {
	ProtocolInterceptor() ServerInterceptor
}
