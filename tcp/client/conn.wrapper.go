package client

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
	"net"
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
	_ = c.conn.Close()
}

func (c *connWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connWrapper) RawWrite(ctx context.Context, message []byte) (n int, err error) {
	return c.conn.Write(message)
}
