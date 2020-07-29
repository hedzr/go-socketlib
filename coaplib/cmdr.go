package coaplib

import (
	"github.com/hedzr/cmdr"
	pi2 "github.com/hedzr/go-socketlib/coaplib/pi"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/go-socketlib/tcp/server"
)

func AttachToCmdr(cmd cmdr.OptCmd, opts ...server.CmdrOpt) {

	// server

	var pi protocol.Interceptor = pi2.NewCoAPInterceptor()
	optx1 := server.WithServerProtocolInterceptor(pi)
	optx2 := server.WithServerPrefixInConfigFile("coaplib.server")
	opt1 := server.WithCmdrServerOptions(optx1, optx2)
	opt2 := server.WithCmdrPort(5688)
	opt3 := server.WithCmdrCommandAction(serverRun)
	opt4 := server.WithCmdrUDPMode(true)

	server.AttachToCmdr(cmd, append(opts, opt1, opt2, opt3, opt4)...)

	serverCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("server"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(serverCmdrOpt)

	// client

	ox1 := client.WithCmdrPort(5688)
	ox2 := client.WithCmdrCommandAction(clientRun)
	ox3 := client.WithCmdrUDPMode(true)

	client.AttachToCmdr(cmd, ox1, ox2, ox3)

	clientCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("client"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(clientCmdrOpt)

}

func serverRun(cmd *cmdr.Command, args []string, opts ...server.Opt) (err error) {
	//var pi server.protocolInterceptor = NewCoAPInterceptor()
	//var opt server.Opt = server.WithServerProtocolInterceptor(pi)
	//return server.DefaultLooper(cmd, args, append(opts, opt)...)
	return server.DefaultLooper(cmd, args, opts...)
}

func clientRun(cmd *cmdr.Command, args []string) (err error) {
	return client.DefaultLooper(cmd, args)
}
