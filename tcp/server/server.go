/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"context"
	"os"

	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
)

func newServer(config *base.Config, preferLogger log.Logger, opts ...Opt) (serve ServeFunc, so *Obj, tlsEnabled bool, err error) {
	so = newServerObj(config)

	for _, opt := range opts {
		opt(so)
	}

	config.UpdatePrefixInConfigFile(so.prefix)
	so.pfs = config.BuildPidFileStruct()
	so.netType = cmdr.GetStringRP(config.PrefixInConfigFile, "network",
		cmdr.GetStringRP(config.PrefixInCommandLine, "network", so.netType))

	if preferLogger != nil {
		config.Logger = preferLogger
	} else {
		config.BuildLogger()
	}
	so.SetLogger(config.Logger)
	if i, ok := so.protocolInterceptor.(interface{ SetLogger(log.Logger) }); ok {
		i.SetLogger(config.Logger)
	}

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

	baseCtx := context.Background()

	switch so.isUDP() {
	case true:
		err = so.createUDPListener(baseCtx, config)
		if err != nil {
			so.Fatalf("build UDP listener failed: %v", err)
		}

		if err = so.pfs.Create(baseCtx); err != nil {
			so.Fatalf("failed to create pid file: %v", err)
		} else {
			so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
		}

	default:
		tlsEnabled, err = so.createListener(baseCtx)
		if err != nil {
			so.Fatalf("build listener failed: %v", err)
		}

		if err = so.pfs.Create(baseCtx); err != nil {
			so.Fatalf("failed to create pid file: %v", err)
		} else {
			so.Infof("PID (%v) file created at: %v", os.Getpid(), so.pfs)
		}

	}

	serve = func(baseCtx context.Context) error {
		return so.Serve(baseCtx)
	}
	return
}

// const prefixSuffix = "server.tls"
const defaultNetType = "tcp"

type CommandAction func(cmd *cmdr.Command, args []string, prefixPrefix string, opts ...Opt) (err error)

func DefaultCommandAction(cmd *cmdr.Command, args []string, prefixPrefix string, opts ...Opt) (err error) {
	var (
		serve      ServeFunc
		so         *Obj
		tlsEnabled bool
		done       = make(chan bool, 1)
	)

	config := base.NewConfigFromCmdrCommand(true, prefixPrefix, cmd)
	serve, so, tlsEnabled, err = newServer(config, cmdr.Logger, opts...)
	if err != nil {
		if so != nil {
			so.Fatalf("build listener failed: %v", err)
		}
		return
	}

	go func() {
		baseCtx := context.WithValue(context.Background(), CTX_SERVER_OBJECT_KEY, so)

		so.protocolInterceptor.OnListened(baseCtx, config.Addr)
		if tlsEnabled {
			so.Printf("Listening on %s with TLS enabled.", config.Addr)
		} else {
			so.Printf("Listening on %s.", config.Addr)
		}

		if err = serve(baseCtx); err != nil {
			if err.Error() == "server closed" {
				err = nil
			} else {
				so.Errorf("Serve() failed: %v", err)
			}
		}
		done <- true // I'm done, cmdr.TrapSignalsEnh should end itself now
	}()

	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		so.Debugf("signal %v caught, requesting shutdown ...", s)
		so.RequestShutdown()
	})()

	return
}
