package pi

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
)

func NewCoAPClientInterceptor() protocol.ClientInterceptor {
	return &pic{}
}

type pic struct {
}

func (p *pic) OnConnected(ctx context.Context, c base.Conn) {
	c.Logger().Debugf("OnConnected")
	return

}

func (p *pic) OnClosing(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosing")
	return

}

func (p *pic) OnClosed(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosed")
	return

}

func (p *pic) OnError(ctx context.Context, c base.Conn, err error) {
	c.Logger().Errorf("OnError: %v", err)
	return

}

func (p *pic) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnReading")
	return

}

func (p *pic) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnWriting")
	return

}

func (p *pic) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	c.Debugf("OnUDPReading")
	return
}

func (p *pic) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	c.Debugf("OnUDPWriting")
	return
}
