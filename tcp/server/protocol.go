package server

import "context"

type ProtocolInterceptor interface {
	OnServerReady(ctx context.Context, s *Obj)
	OnServerClosed(s *Obj)

	OnConnected(ctx context.Context, c Connection)
	OnClosing(c Connection)
	OnClosed(c Connection)

	OnError(ctx context.Context, c Connection, err error)

	OnReading(ctx context.Context, c Connection, data []byte) (processed bool, err error)
	OnWriting(ctx context.Context, c Connection, data []byte) (processed bool, err error)
}
