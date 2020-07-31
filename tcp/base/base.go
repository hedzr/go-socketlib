package base

import (
	"context"
	"github.com/hedzr/log"
	"net"
)

type base struct {
	log.Logger
}

func newBase( /*config *log.LoggerConfig*/ ) base {
	return base{
		log.NewStdLogger(), // build.New(config),
	}
}

type Conn interface {
	Logger() log.Logger

	Close()

	// RawWrite does write through the internal net.Conn
	RawWrite(ctx context.Context, message []byte) (n int, err error)
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
