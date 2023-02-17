package main

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"

	"github.com/hedzr/go-socketlib/_examples/cmdr-simple/opts"
	"github.com/hedzr/go-socketlib/trace"
	// "github.com/hedzr/cmdr-addons/pkg/plugins/trace"
	// _ "github.com/hedzr/logex/logx/zap"
	// _ "github.com/hedzr/logex/logx/zap/sugar"
)

func main() {
	if err := cmdr.Exec(buildRootCmd(),
		cmdr.WithLogx(build.New(log.NewLoggerConfigWith(
			true, "logrus", "debug",
			log.WithTimestamp(true, "")))),
		// cmdr.WithLogex(cmdr.Level(log.WarnLevel)),

		// add '--trace' command-line flag and enable logex.GetTraceMode/cmdr.GetTraceMode
		// trace.WithTraceEnable(true),

		// since cmdr v1.11.10:
		// cmdr.WithIfOpt(isdelve.Enabled, func() cmdr.ExecOption {
		// 	return pprof.GetCmdrProfilingOptions()
		// }),

		// enable '--trace' command line option to toggle a internal trace mode (can be retrieved by cmdr.GetTraceMode())
		trace.WithTraceEnable(defaultTraceEnabled),

		cmdr.WithXrefBuildingHooks(nil, func(root *cmdr.RootCommand, args []string) {
			root.FindSubCommand("generate").Hidden = true
		}),

		// cmdr.WithUnknownOptionHandler(onUnknownOptionHandler),
		// cmdr.WithUnhandledErrorHandler(onUnhandledErrorHandler),
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

	opts.AttachToCmdr(root)
	return
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
	defaultTraceEnabled = true
)
