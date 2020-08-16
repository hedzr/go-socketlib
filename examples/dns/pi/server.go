package pi

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
)

func NewDNSInterceptor() protocol.Interceptor {
	return &dnsServer{}
}

type dnsServer struct {
}

func (s *dnsServer) OnListened(ctx context.Context, c base.Conn) {
	panic("implement me")
}

func (s *dnsServer) OnServerReady(ctx context.Context, c log.Logger) {
	panic("implement me")
}

func (s *dnsServer) OnServerClosed(server log.Logger) {
	panic("implement me")
}

func (s *dnsServer) OnConnected(ctx context.Context, c base.Conn) {
	panic("implement me")
}

func (s *dnsServer) OnClosing(c base.Conn, reason int) {
	panic("implement me")
}

func (s *dnsServer) OnClosed(c base.Conn, reason int) {
	panic("implement me")
}

func (s *dnsServer) OnError(ctx context.Context, c base.Conn, err error) {
	panic("implement me")
}

func (s *dnsServer) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	panic("implement me")
}

func (s *dnsServer) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	panic("implement me")
}

func (s *dnsServer) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	panic("implement me")
}

func (s *dnsServer) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	panic("implement me")
}
