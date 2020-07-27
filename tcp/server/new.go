package server

import "errors"

func New(config *Config, opts ...Opt) ServeFunc {
	return newServer(config, opts...)
}

func WithServerNewConnectionFunc(fn NewConnectionFunc) Opt {
	return func(so *Obj) {
		so.newConnFunc = fn
	}
}

func WithServerProtocolInterceptor(fn ProtocolInterceptor) Opt {
	return func(so *Obj) {
		so.protocolInterceptor = fn
	}
}

func WithServerPrefixInConfigFile(prefix string) Opt {
	return func(so *Obj) {
		so.prefix = prefix
	}
}

type Opt func(so *Obj)
type ServeFunc func() error

var ErrServerClosed = errors.New("server closed")

const (
	DefaultPort = 8883
)
