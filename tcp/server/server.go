/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/log"
	"github.com/hedzr/logex/build"
	"net"
	"os"
	"strconv"
)

func newServer(config *base.Config, opts ...Opt) (serve ServeFunc, so *Obj, tlsEnabled bool, err error) {
	var logger log.Logger
	// logger = build.New(config.LoggerConfig)
	so = newServerObj(logger)

	for _, opt := range opts {
		opt(so)
	}

	so.pfs = makePidFSFromDir(config.PidDir)

	config.PrefixInConfigFile = so.prefix
	so.pfs = makePidFS(config.PrefixInCommandLine)
	so.netType = cmdr.GetStringRP(config.PrefixInConfigFile, "network",
		cmdr.GetStringRP(config.PrefixInCommandLine, "network", so.netType))

	if cmdr.GetBoolRP(config.PrefixInCommandLine, "stop", false) {
		if err = findAndShutdownTheRunningInstance(so.pfs); err != nil {
			so.Errorf("No running instance found: %v", err)
		}
		return
	}

	so.Infof("Starting server... cmdr.InDebugging = %v", cmdr.InDebugging())
	so.Tracef("    logging.level: %v", so.Logger.GetLevel())
	// so.Infof("Starting server...")

	var host, port string
	host, port, err = net.SplitHostPort(config.Addr)
	if port == "" {
		port = strconv.FormatInt(cmdr.GetInt64RP(config.PrefixInConfigFile, "ports.default"), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(config.PrefixInCommandLine, "port", 1024), 10)
		if port == "0" {
			so.Fatalf("invalid port number: %q", port)
		}
	}
	//if host == "" {
	//	host = "0.0.0.0"
	//	// forceIPv6 make all IPv6 ip-addresses of this PC are listened, instead of its IPv4 addresses
	//	const forceIPv6 = false
	//	if forceIPv6 {
	//		host = "[::]"
	//	}
	//}
	config.Addr = net.JoinHostPort(host, port)

	switch so.isUDP() {
	case true:
		err = so.createUDPListener(config)
		if err != nil {
			so.Fatalf("build UDP listener failed: %v", err)
		}

		if err = so.pfs.Create(); err != nil {
			so.Fatalf("failed to create pid file: %v", err)
		} else {
			so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
		}

	default:
		tlsEnabled, err = so.createListener(config)
		if err != nil {
			so.Fatalf("build listener failed: %v", err)
		}

		if err = so.pfs.Create(); err != nil {
			so.Fatalf("failed to create pid file: %v", err)
		} else {
			so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
		}

	}

	serve = so.Serve
	return
}

func DefaultLooper(cmd *cmdr.Command, args []string, opts ...Opt) (err error) {
	loggerConfig := log.NewLoggerConfig()
	_ = cmdr.GetSectionFrom("logger", &loggerConfig)

	config := base.NewConfig()
	config.PrefixInCommandLine = cmd.GetDottedNamePath()
	config.PrefixInConfigFile = "tcp.server"
	config.LoggerConfig = loggerConfig

	var logger log.Logger
	logger = build.New(config.LoggerConfig)
	opts = append(opts, WithServerLogger(logger))

	var (
		serve      ServeFunc
		so         *Obj
		tlsEnabled bool
	)
	serve, so, tlsEnabled, err = newServer(config, opts...)
	if err != nil {
		if so != nil {
			so.Fatalf("build listener failed: %v", err)
		}
		return
	}

	done := make(chan bool, 1)
	go func() {
		defer func() {
			done <- true
		}()
		if tlsEnabled {
			so.Printf("Listening on %s with TLS enabled.", config.Addr)
		} else {
			so.Printf("Listening on %s.", config.Addr)
		}

		baseCtx := context.Background()
		if err = serve(baseCtx); err != nil {
			so.Errorf("Serve() failed: %v", err)
		}
	}()

	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		so.RequestShutdown()
	})()

	return
}
