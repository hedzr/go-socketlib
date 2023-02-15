package base

import (
	"context"
	"net"

	"github.com/hedzr/log"
)

type Conn interface {
	Logger() log.Logger

	Close()

	RemoteAddr() net.Addr

	// RawWrite does write through the internal net.Conn
	RawWrite(ctx context.Context, message []byte) (n int, err error)
}
