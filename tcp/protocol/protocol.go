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

	OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
	OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
}

type Interceptor interface {
	OnServerReady(ctx context.Context, server log.Logger)
	OnServerClosed(server log.Logger)

	ClientInterceptor
}

type ClientInterceptorHolder interface {
	ProtocolInterceptor() ClientInterceptor
}

type InterceptorHolder interface {
	ProtocolInterceptor() Interceptor
}
