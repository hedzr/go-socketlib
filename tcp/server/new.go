package server

import (
	"context"
	"errors"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
)

func New(config *base.Config, opts ...Opt) (serve ServeFunc, err error) {
	serve, _, _, err = newServer(config, opts...)
	return
}

func WithServerNewConnectionFunc(fn NewConnectionFunc) Opt {
	return func(so *Obj) {
		so.newConnFunc = fn
	}
}

func WithServerProtocolInterceptor(fn protocol.Interceptor) Opt {
	return func(so *Obj) {
		so.protocolInterceptor = fn
	}
}

func WithServerPrefixInConfigFile(prefix string) Opt {
	return func(so *Obj) {
		so.prefix = prefix
	}
}

func WithServerUDPMode(mode bool) Opt {
	return func(so *Obj) {
		so.netType = "udp"
	}
}

func WithServerLogger(logger log.Logger) Opt {
	return func(so *Obj) {
		so.Logger = logger
	}
}

// WithServerNetworkType setups the network type of the listener.
//
// The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
//
// For TCP networks, if the host in the address parameter is empty or
// a literal unspecified IP address, Listen listens on all available
// unicast and anycast IP addresses of the local system.
// To only use IPv4, use network "tcp4".
// The address can use a host name, but this is not recommended,
// because it will create a listener for at most one of the host's IP
// addresses.
// If the port in the address parameter is empty or "0", as in
// "127.0.0.1:" or "[::1]:0", a port number is automatically chosen.
// The Addr method of Listener can be used to discover the chosen
// port.
//
func WithServerNetworkType(typeOfNetwork string) Opt {
	return func(so *Obj) {
		so.netType = typeOfNetwork
	}
}

type Opt func(so *Obj)
type ServeFunc func(baseCtx context.Context) error

var ErrServerClosed = errors.New("server closed")

const (
	DefaultPort = 8883
)
