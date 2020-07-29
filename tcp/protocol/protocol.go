package protocol

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
)

type Interceptor interface {
	OnServerReady(ctx context.Context, server log.Logger)
	OnServerClosed(server log.Logger)

	OnConnected(ctx context.Context, c base.Conn)
	OnClosing(c base.Conn)
	OnClosed(c base.Conn)

	OnError(ctx context.Context, c base.Conn, err error)

	OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error)
	OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
	OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error)
}

type InterceptorHolder interface {
	ProtocolInterceptor() Interceptor
}
