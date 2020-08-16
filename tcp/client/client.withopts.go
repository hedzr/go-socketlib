package client

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/log"
	"strings"
)

func New(udpMode bool, config *base.Config, opts ...Opt) error {
	opts = append(opts, WithClientUDPMode(udpMode))
	return defaultLooper(config, nil, nil, "", opts...)
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

func WithClientProtocolInterceptor(fn protocol.ClientInterceptor) Opt {
	return func(obj *clientObj) {
		obj.protocolInterceptor = &ciWrapper{fn}
	}
}

type ciWrapper struct {
	ci protocol.ClientInterceptor
}

func (s *ciWrapper) OnListened(ctx context.Context, c base.Conn) {
	// s.ci.OnListened(ctx, c)
}

func (s *ciWrapper) OnServerReady(ctx context.Context, server log.Logger) {
	panic("implement me")
}

func (s *ciWrapper) OnServerClosed(server log.Logger) {
	panic("implement me")
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
