package udp

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
	"net"
)

type udpConnWrapper struct {
	base.CachedUDPWriter
	conn   *net.UDPConn
	logger log.Logger
}

func (c *udpConnWrapper) Logger() log.Logger {
	return c.logger
}

func (c *udpConnWrapper) Close() {
	_ = c.conn.Close()
}

func (c *udpConnWrapper) RawWrite(ctx context.Context, message []byte) (n int, err error) {
	return c.conn.Write(message)
}

func (c *udpConnWrapper) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *udpConnWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

//type connWrapper struct {
//	*Obj
//}
//
//func (c *connWrapper) Logger() log.Logger {
//	return c.Obj.Logger
//}
//
//func (c *connWrapper) Close() {
//	_ = c.conn.Close()
//}
