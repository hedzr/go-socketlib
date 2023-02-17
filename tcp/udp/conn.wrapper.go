package udp

import (
	"context"
	"net"
	"time"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
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

func (c *udpConnWrapper) WriteNow(message []byte, deadline ...time.Duration) (n int, err error) {
	for _, dur := range deadline {
		err = c.conn.SetWriteDeadline(time.Now().Add(dur))
		if err != nil {
			c.logger.Errorf("error set writing deadline: %v", err)
			return
		}
	}
	return c.conn.Write(message)
}

func (c *udpConnWrapper) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}

func (c *udpConnWrapper) String() string {
	if a := c.conn.RemoteAddr(); a != nil {
		return a.String()
	}
	return c.conn.LocalAddr().String()
}

func (c *udpConnWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (s *udpConnWrapper) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

// type connWrapper struct {
//	*Obj
// }
//
// func (c *connWrapper) Logger() log.Logger {
//	return c.Obj.Logger
// }
//
// func (c *connWrapper) Close() {
//	_ = c.conn.Close()
// }
