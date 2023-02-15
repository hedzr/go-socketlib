package opts

import (
	"context"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
)

func newClientPI() *piClient {
	return &piClient{}
}

type piClient struct {
}

func (p *piClient) OnConnected(ctx context.Context, c base.Conn) {
	// panic("implement me")
}

func (p *piClient) OnClosing(c base.Conn, reason int) {
	// panic("implement me")
}

func (p *piClient) OnClosed(c base.Conn, reason int) {
	// panic("implement me")
}

func (p *piClient) OnError(ctx context.Context, c base.Conn, err error) {
	// panic("implement me")
}

func (p *piClient) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (p *piClient) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (p *piClient) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	// panic("implement me")
	return
}

func (p *piClient) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	// panic("implement me")
	return
}

func (p *piClient) Write() {
	//
}
