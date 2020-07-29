package coaplib

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/protocol"
	"github.com/hedzr/go-socketlib/tcp/server"
)

func AttachToCmdr(cmd cmdr.OptCmd, opts ...server.CmdrOpt) {
	var pi protocol.Interceptor = newCoAPServer()
	optx1 := server.WithServerProtocolInterceptor(pi)
	optx2 := server.WithServerPrefixInConfigFile("coaplib.server")

	opt1 := server.WithCmdrPort(1379)
	opt2 := server.WithCmdrCommandAction(serverRun)
	opt3 := server.WithCmdrServerOptions(optx1, optx2)
	server.AttachToCmdr(cmd, append(opts, opt1, opt2, opt3)...)

	serverCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("server"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(serverCmdrOpt)

	//

	ox1 := client.WithCmdrPort(1379)
	ox2 := client.WithCmdrCommandAction(clientRun)
	client.AttachToCmdr(cmd, ox1, ox2)

	clientCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("client"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(clientCmdrOpt)

}

func serverRun(cmd *cmdr.Command, args []string, opts ...server.Opt) (err error) {
	//var pi server.protocolInterceptor = newCoAPServer()
	//var opt server.Opt = server.WithServerProtocolInterceptor(pi)
	//return server.DefaultLooper(cmd, args, append(opts, opt)...)
	return server.DefaultLooper(cmd, args, opts...)
}

func clientRun(cmd *cmdr.Command, args []string) (err error) {
	return client.DefaultLooper(cmd, args)
}
