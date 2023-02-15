package client

import (
	"context"
	"strings"

	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
)

func New(udpMode bool, config *base.Config, opts ...Opt) error {
	// opts = append(opts, WithClientUDPMode(udpMode))
	if udpMode {
		config.Network = "udp"
	}
	return DefaultCommandAction(config, nil, opts...)
}

type Opt func(obj *clientObj)

func WithClientUDPMode(udpMode bool) Opt {
	return func(obj *clientObj) {
		if udpMode {
			obj.mainLoop = defaultUdpMainLoop
		} else {
			obj.mainLoop = defaultMainLoop
		}
	}
}

func WithClientMainLoop(mainLoop MainLoop) Opt {
	return func(obj *clientObj) {
		obj.mainLoop = mainLoop
	}
}

func WithClientPrefixPrefix(prefixPrefix string) Opt {
	return func(obj *clientObj) {
		obj.prefixInConfigFile = strings.Join([]string{prefixPrefix, "client", "tls"}, ".")
	}
}

func WithClientProtocolInterceptor(pic protocol.ClientInterceptor) Opt {
	return func(obj *clientObj) {
		obj.protocolInterceptor = &ciWrapper{pic}
		if mlh, ok := pic.(MainLoopHolder); ok {
			obj.mainLoop = mlh.MainLoop
		}
	}
}

func WithClientBuildPackageFunc(fn BuildPackageFunc) Opt {
	return func(obj *clientObj) {
		obj.buildPackage = fn
	}
}

type ciWrapper struct {
	ci protocol.ClientInterceptor
}

func (s *ciWrapper) OnListened(baseCtx context.Context, addr string) {
	// panic("implement me")
}

func (s *ciWrapper) OnServerReady(ctx context.Context, server log.Logger) {
	// panic("implement me")
}

func (s *ciWrapper) OnServerClosed(server log.Logger) {
	// panic("implement me")
}

func (s *ciWrapper) OnConnected(ctx context.Context, c base.Conn) {
	s.ci.OnConnected(ctx, c)
}

func (s *ciWrapper) OnClosing(c base.Conn, reason int) {
	s.ci.OnClosing(c, reason)
}

func (s *ciWrapper) OnClosed(c base.Conn, reason int) {
	s.ci.OnClosed(c, reason)
}

func (s *ciWrapper) OnError(ctx context.Context, c base.Conn, err error) {
	s.ci.OnError(ctx, c, err)
}

func (s *ciWrapper) OnReading(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return s.ci.OnReading(ctx, c, data)
}

func (s *ciWrapper) OnWriting(ctx context.Context, c base.Conn, data []byte) (processed bool, err error) {
	return s.ci.OnWriting(ctx, c, data)
}

func (s *ciWrapper) OnUDPReading(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return s.ci.OnUDPReading(ctx, c, packet)
}

func (s *ciWrapper) OnUDPWriting(ctx context.Context, c log.Logger, packet *base.UdpPacket) (processed bool, err error) {
	return s.ci.OnUDPWriting(ctx, c, packet)
}
