package pi

import (
	"context"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
)

func NewDNSClientInterceptor() protocol.ClientInterceptor {
	return &dnsClient{}
}

type dnsClient struct {
}

func (s *dnsClient) OnConnected(ctx context.Context, c base.Conn) {
	// panic("implement me")
}

func (s *dnsClient) OnClosing(c base.Conn, reason int) {

}

func (s *dnsClient) OnClosed(c base.Conn, reason int) {

}

func (s *dnsClient) OnError(ctx context.Context, c base.Conn, err error) {

}

func (s *dnsClient) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (s *dnsClient) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (s *dnsClient) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return
}

func (s *dnsClient) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return
}
