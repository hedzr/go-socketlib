/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/logex/build"
	"net"
	"os"
	"strconv"
)

func newServer(config *Config, opts ...Opt) ServeFunc {
	so := newServerObj(build.New(config.LoggerConfig))

	for _, opt := range opts {
		opt(so)
	}

	so.pfs = makePidFSFromDir(config.PidDir)

	so.Infof("Starting server...")

	var addr, host, port string
	var err error
	host, port, err = net.SplitHostPort(config.Addr)
	addr = net.JoinHostPort(host, port)

	var listener net.Listener
	var tlsEnabled bool
	listener, tlsEnabled, err = serverBuildListener(so, addr, config.PrefixInConfigFile, config.PrefixInCommandLine)
	if err != nil {
		so.Fatalf("build listener failed: %v", err)
	}
	so.Listen(listener)

	if err = so.pfs.Create(); err != nil {
		so.Fatalf("failed to create pid file: %v", err)
	} else {
		so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
	}

	if tlsEnabled {
		so.Printf("Listening on %s with TLS enabled.", addr)
	} else {
		so.Printf("Listening on %s.", addr)
	}

	return so.Serve
}

func DefaultLooper(cmd *cmdr.Command, args []string, opts ...Opt) (err error) {

	loggerConfig := build.NewLoggerConfig()
	_ = cmdr.GetSectionFrom("logger", &loggerConfig)

	config := NewConfig()
	config.PrefixInCommandLine = cmd.GetDottedNamePath()
	config.PrefixInConfigFile = "tcp.server"
	config.LoggerConfig = loggerConfig

	so := newServerObj(build.New(loggerConfig))

	for _, opt := range opts {
		opt(so)
	}

	config.PrefixInConfigFile = so.prefix
	so.pfs = makePidFS(config.PrefixInCommandLine)

	if cmdr.GetBoolRP(config.PrefixInCommandLine, "stop", false) {
		if err = findAndShutdownTheRunningInstance(so.pfs); err != nil {
			so.Errorf("No running instance found: %v", err)
		}
		return
	}

	so.Infof("Starting server... cmdr.InDebugging = %v", cmdr.InDebugging())
	so.Tracef("    logging.level: %v", so.Logger.GetLevel())

	var addr, host, port string
	host, port, err = net.SplitHostPort(cmdr.GetStringRP(config.PrefixInConfigFile, "addr"))
	if port == "" {
		port = strconv.FormatInt(cmdr.GetInt64RP(config.PrefixInConfigFile, "ports.default"), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(config.PrefixInCommandLine, "port", 1024), 10)
		if port == "0" {
			so.Fatalf("invalid port number: %q", port)
		}
	}
	addr = net.JoinHostPort(host, port)

	var listener net.Listener
	var tlsEnabled bool
	listener, tlsEnabled, err = serverBuildListener(so, addr, config.PrefixInConfigFile, config.PrefixInCommandLine)
	if err != nil {
		so.Fatalf("build listener failed: %v", err)
	}
	so.Listen(listener)

	if err = so.pfs.Create(); err != nil {
		so.Fatalf("failed to create pid file: %v", err)
	} else {
		so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
	}

	go func() {
		if tlsEnabled {
			so.Printf("Listening on %s with TLS enabled.", addr)
		} else {
			so.Printf("Listening on %s.", addr)
		}
		if err = so.Serve(); err != nil {
			so.Errorf("Serve() failed: %v", err)
		}
	}()

	tcp.HandleSignals(func(s os.Signal) {
		so.RequestShutdown()
	})()
	return
}

func serverBuildListener(so *Obj, addr, prefixInConfigFile, prefixInCommandLine string) (listener net.Listener, tls bool, err error) {
	var tlsListener net.Listener
	listener, err = net.Listen(
		cmdr.GetStringRP(prefixInConfigFile, "network",
			cmdr.GetStringRP(prefixInCommandLine, "network", "tcp")),
		addr)
	if err != nil {
		so.Fatalf("error: %v", err)
	}

	ctcPrefix := prefixInConfigFile + ".tls"
	ctc := tls2.NewCmdrTlsConfig(ctcPrefix, prefixInCommandLine)
	so.Debugf("%v", ctc)
	if ctc.Enabled {
		tlsListener, err = ctc.NewTlsListener(listener)
		if err != nil {
			so.Fatalf("error: %v", err)
		}
	}
	if tlsListener != nil {
		listener = tlsListener
		tls = true
	}
	return
}
