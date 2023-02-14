package opts

import (
	"context"
	"github.com/hedzr/go-socketlib/_examples/cmdr/pibufio"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
	"sync/atomic"
	"time"
)

func newServerPI() *piServer {
	chExit := make(chan struct{})
	r := &piServer{
		chExit: chExit,
		iobuf:  pibufio.New(),
	}
	go r.run()
	return r
}

type piServer struct {
	chExit        chan struct{}
	iobuf         pibufio.Queue
	totalReceived int64
}

func (p *piServer) Close() {
	close(p.chExit)
}

func (p *piServer) run() {
	ticker := time.NewTicker(777 * time.Microsecond)
	defer func() {
		ticker.Stop()
	}()

	//if o, ok := p.iobuf.(interface{ SetTraceEnabled(b bool) }); ok {
	//	b := log.GetTraceMode() || log.GetDebugMode()
	//	o.SetTraceEnabled(b)
	//}

	for {
		select {
		case pkg := <-p.iobuf.PkgReceived():
			p.onPkgReceived(pkg)
		case <-ticker.C:
			p.iobuf.TryExtractPackages()
		case <-p.chExit:
			return
		}
	}
}

func (p *piServer) OnListened(baseCtx context.Context, addr string) {
	//panic("implement me")
	log.Debugf("(pi) onListened at %v...", addr)
}

func (p *piServer) OnServerReady(ctx context.Context, c log.Logger) {
	//panic("implement me")
	c.Debugf("(pi) onServerReady...")
}

func (p *piServer) OnServerClosed(server log.Logger) {
	//panic("implement me")
	server.Debugf("(pi) onServerClosed...")
}

func (p *piServer) OnConnected(ctx context.Context, c base.Conn) {
	//panic("implement me")
	c.Logger().Debugf("(pi) onConnected...")
}

func (p *piServer) OnClosing(c base.Conn, reason int) {
	//panic("implement me")
	c.Logger().Debugf("(pi) onClosing...(reason=%v)", reason)
}

func (p *piServer) OnClosed(c base.Conn, reason int) {
	//panic("implement me")
	c.Logger().Debugf("(pi) onClosed...(reason=%v)", reason)
}

func (p *piServer) OnError(ctx context.Context, c base.Conn, err error) {
	//panic("implement me")
	c.Logger().Debugf("(pi) onError: %v", err)
}

func (p *piServer) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	err = p.iobuf.Enqueue(data)
	return
}

func (p *piServer) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	//panic("implement me")
	return
}

func (p *piServer) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	//panic("implement me")
	return
}

func (p *piServer) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	//panic("implement me")
	return
}

func (p *piServer) onPkgReceived(pkg []byte) {
	atomic.AddInt64(&p.totalReceived, 1)
	//log.Debugf("pkg-received (#%v): %v", p.totalReceived, pkg)
}
