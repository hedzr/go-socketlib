package main

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/examples/dns/pi"
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

	// DNS server and client

	tcpCmd := root.NewSubCommand("dns", "dns").
		Description("DNS Server/Client operations...", "")

	var pis = pi.NewDNSInterceptor()
	opt1 := server.WithCmdrServerProtocolInterceptor(pis)
	opt2 := server.WithCmdrPort(60053)
	opt5 := server.WithCmdrPrefixPrefix("dns")

	server.AttachToCmdr(tcpCmd, opt1, opt2, opt5)

	var pic = pi.NewDNSClientInterceptor()
	ox1 := client.WithCmdrClientProtocolInterceptor(pic)
	ox2 := client.WithCmdrPort(53)            // get ports configs from config file
	ox6 := client.WithCmdrPrefixPrefix("dns") // prefix of prefix is used for loading the coap section from config file

	client.AttachToCmdr(tcpCmd, ox1, ox2, ox6)

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
