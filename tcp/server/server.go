/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/base"
	"os"
)

func newServer(config *base.Config, opts ...Opt) (serve ServeFunc, so *Obj, tlsEnabled bool, err error) {
	// var logger log.Logger
	// logger = build.New(config.LoggerConfig)
	so = newServerObj()

	for _, opt := range opts {
		opt(so)
	}

	config.UpdatePrefixInConfigFile(so.prefix)
	so.pfs = config.BuildPidFile()
	so.netType = cmdr.GetStringRP(config.PrefixInConfigFile, "network",
		cmdr.GetStringRP(config.PrefixInCommandLine, "network", so.netType))

	config.BuildLogger()
	so.SetLogger(config.Logger)

	if cmdr.GetBoolRP(config.PrefixInCommandLine, "stop", false) {
		if err = base.FindAndShutdownTheRunningInstance(so.pfs); err != nil {
			so.Errorf("No running instance found: %v", err)
		}
		return
	}

	so.Infof("Starting server (%v)... cmdr.InDebugging = %v", so.netType, cmdr.InDebugging())
	so.Tracef("    logging.level: %v", so.Logger.GetLevel())
	// so.Infof("Starting server...")

	if err = config.BuildServerAddr(); err != nil {
		config.Logger.Fatalf("%v", err)
	}

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

//const prefixSuffix = "server.tls"
const defaultNetType = "tcp"

type CommandAction func(cmd *cmdr.Command, args []string, prefixPrefix string, opts ...Opt) (err error)

func DefaultLooper(cmd *cmdr.Command, args []string, prefixPrefix string, opts ...Opt) (err error) {
	var (
		serve      ServeFunc
		so         *Obj
		tlsEnabled bool
	)
	config := base.NewConfigFromCmdrCommand(true, prefixPrefix, cmd)
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
		so.Debugf("signal %v caught, requesting shutdown ...")
		so.RequestShutdown()
	})()

	return
}
