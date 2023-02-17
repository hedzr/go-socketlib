package base

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/hedzr/log"
)

type Conn interface {
	Logger() log.Logger

	Close()

	RemoteAddr() net.Addr
	LocalAddr() net.Addr

	// RawWrite does write through the internal net.Conn
	RawWrite(ctx context.Context, message []byte) (n int, err error)

	// WriteNow does write message immediately.
	WriteNow(msg []byte, deadline ...time.Duration) (n int, err error)

	io.Reader
}

type Conn2 interface {
	Conn
}
