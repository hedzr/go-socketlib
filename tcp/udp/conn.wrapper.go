package udp

import (
	"context"
	"github.com/hedzr/log"
	"net"
)

type udpConnWrapper struct {
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
