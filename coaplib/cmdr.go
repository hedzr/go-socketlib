package coaplib

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"
)

func AttachToCmdr(cmd cmdr.OptCmd, opts ...server.CmdrOpt) {
	var pi server.ProtocolInterceptor = newCOAPServer()
	optx1 := server.WithServerProtocolInterceptor(pi)
	optx2 := server.WithServerPrefixInConfigFile("coaplib.server")

	opt1 := server.WithCmdrPort(1379)
	opt2 := server.WithCmdrCommandAction(serverRun)
	opt3 := server.WithCmdrServerOptions(optx1, optx2)
	server.AttachToCmdr(cmd, append(opts, opt1, opt2, opt3)...)

	ox1 := client.WithCmdrPort(1379)
	ox2 := client.WithCmdrCommandAction(clientRun)
	client.AttachToCmdr(cmd, ox1, ox2)
}

func serverRun(cmd *cmdr.Command, args []string, opts ...server.Opt) (err error) {
	//var pi server.ProtocolInterceptor = newCOAPServer()
	//var opt server.Opt = server.WithServerProtocolInterceptor(pi)
	//return server.DefaultLooper(cmd, args, append(opts, opt)...)
	return server.DefaultLooper(cmd, args, opts...)
}

func clientRun(cmd *cmdr.Command, args []string) (err error) {
	return client.DefaultLooper(cmd, args)
}

func newCOAPServer() server.ProtocolInterceptor {
	return &piCOAP{}
}

type piCOAP struct {
}

func (s *piCOAP) OnServerReady(ctx context.Context, so *server.Obj) {
	so.Debugf("OnServerReady")
}

func (s *piCOAP) OnServerClosed(so *server.Obj) {
	so.Debugf("OnServerClosed")
}

func (s *piCOAP) OnConnected(ctx context.Context, c server.Connection) {
	c.Logger().Debugf("OnConnected")
}

func (s *piCOAP) OnClosing(c server.Connection) {
	c.Logger().Debugf("OnClosing")
}

func (s *piCOAP) OnClosed(c server.Connection) {
	c.Logger().Debugf("OnClosed")
}

func (s *piCOAP) OnError(ctx context.Context, c server.Connection, err error) {
	c.Logger().Errorf("OnError: %v", err)
}

func (s *piCOAP) OnReading(ctx context.Context, c server.Connection, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnReading")
	return
}

func (s *piCOAP) OnWriting(ctx context.Context, c server.Connection, data []byte) (processed bool, err error) {
	c.Logger().Debugf("OnWriting")
	return
}
