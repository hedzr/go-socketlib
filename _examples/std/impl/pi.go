package impl

import (
	"context"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
)

func NewServerInterceptor() *serverPI {
	return &serverPI{}
}

type serverPI struct {
	log.Logger
}

func (pi *serverPI) SetLogger(logger log.Logger) {
	pi.Logger = logger
}

func (pi *serverPI) OnListened(ctx context.Context, addr string) {

}

func (pi *serverPI) OnServerReady(ctx context.Context, c log.Logger) {

}

func (pi *serverPI) OnServerClosed(server log.Logger) {

}

func (pi *serverPI) OnConnected(ctx context.Context, c base.Conn) {

}

func (pi *serverPI) OnClosing(c base.Conn, reason int) {

}

func (pi *serverPI) OnClosed(c base.Conn, reason int) {

}

func (pi *serverPI) OnError(ctx context.Context, c base.Conn, err error) {

}

func (pi *serverPI) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (pi *serverPI) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return
}

func (pi *serverPI) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return
}

func (pi *serverPI) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return
}
