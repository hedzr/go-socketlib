package server

import "errors"

func New(config *Config, opts ...Opt) ServeFunc {
	return newServer(config, opts...)
}

func WithNewConnectionFunc(fn NewConnectionFunc) Opt {
	return func(so *Obj) {
		so.newConnFunc = fn
	}
}

type Opt func(so *Obj)
type ServeFunc func() error

var ErrServerClosed = errors.New("server closed")

const (
	DefaultPort = 8883
)
