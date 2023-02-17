package client

import (
	"context"
	"net"
	"time"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
)

type connWrapper struct {
	base.CachedTCPWriter
	conn   net.Conn
	logger log.Logger
}

func (c *connWrapper) Logger() log.Logger {
	return c.logger
}

func (c *connWrapper) Close() {
	c.logger.Debugf("closing c.conn")
	_ = c.conn.Close()
}

func (c *connWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (s *connWrapper) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (c *connWrapper) RawWrite(ctx context.Context, message []byte) (n int, err error) {
	return c.conn.Write(message)
}

func (c *connWrapper) WriteNow(message []byte, deadline ...time.Duration) (n int, err error) {
	for _, dur := range deadline {
		err = c.conn.SetWriteDeadline(time.Now().Add(dur))
		if err != nil {
			c.logger.Errorf("error set writing deadline: %v", err)
			return
		}
	}
	return c.conn.Write(message)
}

func (c *connWrapper) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}
