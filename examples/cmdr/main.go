package main

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/cert"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"

	"github.com/hedzr/cmdr-addons/pkg/plugins/trace"

	"github.com/hedzr/log"
	"github.com/hedzr/logex/logx/logrus"
	_ "github.com/hedzr/logex/logx/zap"
	_ "github.com/hedzr/logex/logx/zap/sugar"
)

func main() {
	if err := cmdr.Exec(buildRootCmd(),
		cmdr.WithLogx(logrus.New("debug", false, true)),
		cmdr.WithLogex(cmdr.Level(log.WarnLevel)),

		// add '--trace' command-line flag and enable logex.GetTraceMode/cmdr.GetTraceMode
		trace.WithTraceEnable(true),
		cmdr.WithXrefBuildingHooks(nil, func(root *cmdr.RootCommand, args []string) {
			root.FindSubCommand("generate").Hidden = true
		}),

		//cmdr.WithUnknownOptionHandler(onUnknownOptionHandler),
		//cmdr.WithUnhandledErrorHandler(onUnhandledErrorHandler),
	); err != nil {
		cmdr.Logger.Fatalf("error: %+v", err)
	}
}

func buildRootCmd() (rootCmd *cmdr.RootCommand) {
	root := cmdr.Root(appName, "1.0.1").
		Copyright(copyright, "hedzr").
		Description(desc, longDesc).
		Examples(examples)
	rootCmd = root.RootCommand()

	socketLibCmd(root)
	return
}

func socketLibCmd(root cmdr.OptCmd) {

	// TCP/UDP

	tcpCmd := root.NewSubCommand("tcp", "tcp", "socket", "socketlib").
		Description("go-socketlib TCO operations...", "").
		Group("Socket")

	server.AttachToCmdr(tcpCmd, server.WithCmdrPort(1983))
	client.AttachToCmdr(tcpCmd, client.WithCmdrPort(1983), client.WithCmdrInteractiveCommand(true))

	udpCmd := root.NewSubCommand("udp", "udp").
		Description("go-socketlib UDP operations...", "").
		Group("Socket")

	server.AttachToCmdr(udpCmd, server.WithCmdrUDPMode(true), server.WithCmdrPort(1984))
	client.AttachToCmdr(udpCmd, client.WithCmdrUDPMode(true), client.WithCmdrPort(1984))

	// Cert

	cert.AttachToCmdr(root)

}

const (
	appName   = "tcp-tool"
	copyright = "tcp-tool is an effective devops tool"
	desc      = "tcp-tool is an effective devops tool. It make an demo application for `cmdr`."
	longDesc  = "tcp-tool is an effective devops tool. It make an demo application for `cmdr`."
	examples  = `
$ {{.AppName}} --help
  show help screen.
`
)
