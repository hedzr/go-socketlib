package pi

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
)

func NewCoAPInterceptor() protocol.Interceptor {
	return &piCoAP{}
}

type piCoAP struct {
}

func (s *piCoAP) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	c.Debugf("OnUDPReading")
	return
}

func (s *piCoAP) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	c.Debugf("OnUDPWriting")
	return
}

func (s *piCoAP) OnServerReady(ctx context.Context, so log.Logger) {
	so.Debugf("OnServerReady")
}

func (s *piCoAP) OnServerClosed(so log.Logger) {
	so.Debugf("OnServerClosed")
}

func (s *piCoAP) OnConnected(ctx context.Context, c base.Conn) {
	c.Logger().Debugf("OnConnected")
}

func (s *piCoAP) OnClosing(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosing")
}

func (s *piCoAP) OnClosed(c base.Conn, reason int) {
	c.Logger().Debugf("OnClosed")
}

func (s *piCoAP) OnError(ctx context.Context, c base.Conn, err error) {
	c.Logger().Errorf("OnError: %v", err)
}

func (s *piCoAP) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnReading")
	return
}

func (s *piCoAP) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnWriting")
	return
}
