# go-socketlib

![Go](https://github.com/hedzr/go-socketlib/workflows/Go/badge.svg)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/hedzr/go-socketlib.svg?label=release)](https://github.com/hedzr/go-socketlib/releases)
[![Go Dev](https://img.shields.io/badge/go-dev-green)](https://pkg.go.dev/github.com/hedzr/go-socketlib)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/hedzr/go-socketlib)
[![Go Report Card](https://goreportcard.com/badge/github.com/hedzr/go-socketlib)](https://goreportcard.com/report/github.com/hedzr/go-socketlib)
[![Coverage Status](https://coveralls.io/repos/github/hedzr/go-socketlib/badge.svg?branch=master&.9)](https://coveralls.io/github/hedzr/go-socketlib?branch=master)

`go-socketlib` provides a simple, fast approach to implement your communication protocol.

## News

WIP. Here is a pre-release version.

- v0.5.2
  - upgrade deps

- v0.5.1
  - security updates

- v0.5.0
  - planning review API

- v0.3.0
  - BREAK: go 1.16.x and below is no longer supported.
  - fixed bugs found.

- v0.2.5
  - old release

## Features

- supports TCP, UDP, and Unix socket server/client developing
- Write your business logical
  with [protocol.Interceptor](https://github.com/hedzr/go-socketlib/blob/master/tcp/protocol/protocol.go#L22)

## Getting Start

### Import

```go
import "github.com/hedzr/go-socketlib"
```

### Write a TCP server

#### With [hedzr/cmdr](https://github.com/hedzr/cmdr)

`go-socketlib` has been integrated with `cmdr`. Here is a full app:

```go
package main

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
)

func main() {
	if err := cmdr.Exec(buildRootCmd(),
		cmdr.WithLogx(build.New(log.NewLoggerConfigWith(true, "sugar", "debug"))),
		//cmdr.WithUnknownOptionHandler(onUnknownOptionHandler),
		//cmdr.WithUnhandledErrorHandler(onUnhandledErrorHandler),
	); err != nil {
		cmdr.Logger.Fatalf("error: %+v", err)
	}
}

func buildRootCmd() (rootCmd *cmdr.RootCommand) {
	root := cmdr.Root(appName, "1.0.1").
		Header("fluent - test for cmdr - no version - hedzr").
		Description(desc, longDesc).
		Examples(examples)
	rootCmd = root.RootCommand()

	socketlibCmd(root)
	return
}

func socketlibCmd(root cmdr.OptCmd) {
	// for TCP server/client
  
	aCmd := root.NewSubCommand("tcp", "tcp", "socket", "socketlib").
		Description("go-socketlib operations...", "").
		Group("TCP")
	server.AttachToCmdr(aCmd, server.WithPort(1983))
	client.AttachToCmdr(aCmd, client.WithCmdrPort(1983))
  
	// for UDP server/client
  
	udpCmd := root.NewSubCommand("udp", "udp").
		Description("go-socketlib UDP operations...", "").
		Group("UDP")

	server.AttachToCmdr(udpCmd, server.WithCmdrUDPMode(true), server.WithCmdrPort(1984))
	client.AttachToCmdr(udpCmd, client.WithCmdrUDPMode(true), client.WithCmdrPort(1984))
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
```

Run it:

```bash
$ go run ./cli tcp --help
tcp-tool is an effective devops tool by hedzr - v1.0.1

Usages:
    tcp-tool tcp [Sub-Commands] [tail args...] [Options] [Parent/Global Options]

Description:
    go-socketlib TCO operations...

Sub-Commands:
  c, client                                       TCP/UDP/Unix client operations
  s, server                                       TCP/UDP/Unix Server Operations

Global Options:
  [Misc]
          --config=[Locations of config files]    load config files from where you specified (default [Locations of config files]=)
  -q,     --quiet                                 No more screen output. [env: QUITE] (default=false)
  -trace, --trace,--tr                            enable trace mode for tcp/mqtt send/recv data dump [env: TRACE] (default=false)
  -v,     --verbose                               Show this help screen [env: VERBOSE] (default=false)

Type '-h'/'-?' or '--help' to get command help screen.
More: '-D'/'--debug'['--env'|'--raw'|'--more'], '-V'/'--version', '-#'/'--build-info', '--no-color', '--strict-mode', '--no-env-overrides'...

```

The server options:

```bash
❯ go run ./cli tcp server --help
tcp-tool is an effective devops tool by hedzr - v1.0.1

Usages:
    tcp-tool tcp server [tail args...] [Options] [Parent/Global Options]

Description:
    TCP/UDP/Unix Server Operations

Options:
     --network                                    network: tcp, tcp4, tcp6, unix, unixpacket, and udp, udp4, udp6 (default='tcp')
  [TCP/UDP]
  -a, --addr=HOST-or-IP,--adr,--address           The address to listen to (default HOST-or-IP=)
  -p, --port=PORT                                 The port to listen on (default PORT=1983)
  [TLS]
  -tls, --enable-tls                              enable TLS mode (default=false)
  -ca,  --cacert=PATH,--ca-cert                   CA cert path (.cer,.crt,.pem) if it's standalone (default PATH='root.pem')
  -c,   --cert=PATH                               server public-cert path (.cer,.crt,.pem) (default PATH='cert.pem')
  -k,   --key=PATH                                server private-key path (.cer,.crt,.pem) (default PATH='cert.key')
        --client-auth                             enable client cert authentication (default=false)
        --tls-version                             tls-version: 0,1,2,3 (default=2)
  [Tool]
  -pp, --pid-path=PATH                            The pid filepath (default PATH='/var/run/$APPNAME/$APPNAME.pid')
  -s,  --stop,--shutdown                          stop/shutdown the running server (default=false)

Global Options:
  [Misc]
          --config=[Locations of config files]    load config files from where you specified (default [Locations of config files]=)
  -q,     --quiet                                 No more screen output. [env: QUITE] (default=false)
  -trace, --trace,--tr                            enable trace mode for tcp/mqtt send/recv data dump [env: TRACE] (default=false)
  -v,     --verbose                               Show this help screen [env: VERBOSE] (default=false)

Type '-h'/'-?' or '--help' to get command help screen.
More: '-D'/'--debug'['--env'|'--raw'|'--more'], '-V'/'--version', '-#'/'--build-info', '--no-color', '--strict-mode', '--no-env-overrides'...

```

start the server:

```bash
❯ go run ./cli tcp server -tls
2020-08-14T17:16:53.782+0800	INFO	go-socketlib/tcp/server/server.go:38	Starting server (tcp)... cmdr.InDebugging = false
2020-08-14T17:16:53.793+0800	INFO	go-socketlib/tcp/server/server.go:70	PID (79656) file created at: /var/run/tcp-tool/tcp-tool.pid
2020-08-14T17:16:53.793+0800	INFO	go-socketlib/tcp/server/server.go:106	Listening on :1983 with TLS enabled.
```

Start the client:

```bash
$ go run ./cli/c2 tcp client -tls -k
...
2020-08-14T17:32:25.279+0800	DEBUG	go-socketlib/tcp/client/client.go:129	 #99 sent
```

### Write the server/client protocol interceptor with yours

By using [Interceptor](https://github.com/hedzr/go-socketlib/blob/master/tcp/protocol/protocol.go#L22) and
ClientInterceptor, you can attach a protocol interceptor onto bare metal socketlib server/client.

For example (from our CoAP impl):

```go
func AttachToCmdr(cmd cmdr.OptCmd, opts ...server.CmdrOpt) {

	// server

	var pis = pi.NewCoAPInterceptor()
	// var pisOpt server.Opt
	// pisOpt = server.WithServerProtocolInterceptor(pis)
	// opt1 := server.WithCmdrServerOptions(pisOpt)
	opt1 := server.WithCmdrServerProtocolInterceptor(pis)
	opt2 := server.WithCmdrPort(0)
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

	var pic = pi.NewCoAPClientInterceptor()
	//optCx1 := client.WithClientProtocolInterceptor(pic)
	//ox1 := client.WithCmdrClientOptions(optCx1)
	ox1 := client.WithCmdrClientProtocolInterceptor(pic)
	ox2 := client.WithCmdrPort(0)                                        // get ports configs from config file
	ox3 := client.WithCmdrUDPMode(true)                                  // enable udp mode and loop (udpLoop)
	ox4 := client.WithCmdrCommandAction(client.DefaultLooper)            // default internal looper
	ox5 := client.WithCmdrMainLoop(pic.(client.MainLoopHolder).MainLoop) // coapMainLoop will block the main thread to exit to OS
	ox6 := client.WithCmdrPrefixPrefix("coap")                           // prefix of prefix is used for loading the coap section from config file

	client.AttachToCmdr(cmd, ox1, ox2, ox3, ox4, ox5, ox6)

	clientCmdrOpt := cmdr.NewCmdFrom(cmd.ToCommand().FindSubCommand("client"))
	cmdr.NewBool().
		Titles("dry-run", "dr", "dryrun").
		Description("In dry-run mode, arguments will be parsed, tcp listener will not be stared.").
		Group("zzz1.Testers").
		AttachTo(clientCmdrOpt)
	cmdr.NewBool().
		Titles("try-debug", "try").
		Description("In try-debug mode, A continuous send/recv (to coap.me) processing will be started automatically.").
		Group("zzz1.Testers").
		AttachTo(clientCmdrOpt)
	clientCmdrOpt.ToCommand().FindFlag("host").DefaultValue = remoteCaliforniumEclipseOrg // remoteCaliforniumEclipseOrg
}
```

## Contrib

Welcome

## LICENSE

MIT
