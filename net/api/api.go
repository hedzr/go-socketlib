package api

import (
	"context"
	"net"
	"time"

	"github.com/hedzr/is/basics"
)

type Conn interface {
	// Logger() log.Logger

	Response
	RawWriteable

	// io.Reader
}

type Request interface {
	Read(p []byte) (n int, err error)

	Addressable
}

type Response interface {
	basics.Closable

	// // Close closes the connection.
	// Close() error

	Addressable
	Writeable
}

type Addressable interface {
	// LocalAddr returns the net.Addr of the server
	LocalAddr() net.Addr
	// RemoteAddr returns the net.Addr of the client that sent the current request.
	RemoteAddr() net.Addr
	RemoteAddrString() string // safe getter for remote address
}

// Writeable provides cacheable writing feature
type Writeable interface {
	// Write writes a raw buffer back to the client.
	Write(data []byte) (n int, err error)
}

// RawWriteable provides instance writing feature without cache.
type RawWriteable interface {
	// RawWrite does write through the internal net.Conn
	RawWrite(ctx context.Context, message []byte) (n int, err error)

	// RawWriteTimeout does write message immediately.
	RawWriteTimeout(msg []byte, deadline ...time.Duration) (n int, err error)
}

type UdpPacket struct {
	RemoteAddr *net.UDPAddr
	Data       []byte
}

func NewUdpPacket(remoteAddr *net.UDPAddr, data []byte) *UdpPacket {
	return &UdpPacket{
		RemoteAddr: remoteAddr,
		Data:       data,
	}
}

type CachedTCPWriter interface {
	// WriteString send the string to the writing queue
	WriteString(message string)
	// Write send the buffer to the writing queue
	Write(message []byte)
}

type CachedUDPWriter interface {
	WriteTo(remoteAddr *net.UDPAddr, data []byte)
}
