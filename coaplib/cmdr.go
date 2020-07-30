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
	opt1 := server.WithCmdrServerOptions(optx1)
	opt2 := server.WithCmdrPort(5688)
	opt3 := server.WithCmdrCommandAction(server.DefaultLooper)
	opt4 := server.WithCmdrUDPMode(true)
	opt5 := server.WithCmdrPrefixPrefix("coap")

	server.AttachToCmdr(cmd, append(opts, opt1, opt2, opt3, opt4, opt5)...)

	serverCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("server"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(serverCmdrOpt)

	// client

	ox1 := client.WithCmdrPort(5688)
	optcx1 := client.WithClientProtocolInterceptor(pi)
	ox2 := client.WithCmdrClientOptions(optcx1)
	ox2 := client.WithCmdrUDPMode(true)
	ox3 := client.WithCmdrCommandAction(client.DefaultLooper)
	ox5 := client.WithCmdrPrefixPrefix("coap")

	client.AttachToCmdr(cmd, ox1, ox2, ox3, ox5)

	clientCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("client"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Dry Run").
		AttachTo(clientCmdrOpt)

}
