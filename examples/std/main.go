package main

import (
	"context"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/server"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/log"
	"os"
)

func main() {

	// app-name, these component need it: pid-file, log-file, ...
	_ = os.Setenv("APPNAME", "std-server")

	// default is zap sugar logger
	logConfig := log.NewLoggerConfig()
	logConfig.Backend = "logrus" // zap, sugar, std, off/dummy/none

	var ignoredKey, ignoredAdapterName string
	config := base.NewConfigWithParams(true,
		"tcp",      // tcp, udp, or unix
		ignoredKey, // ignore safely because you give up from cmdr
		ignoredKey, // ignore safely because you give up from cmdr
		logConfig,
		func(cfg *tls2.CmdrTlsConfig) {
			cfg.Cert = "unknown" // give a valid path if you like to enable TLS
			cfg.Key = "unknown"
		},
		":1983",
		"my-protocol://localhost:1983",
		ignoredAdapterName,
	)

	serve, _, tlsEnabled, err := server.New(config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if tlsEnabled {
		log.Printf("Listening on %s with TLS enabled.", config.Addr)
	} else {
		log.Printf("Listening on %s.", config.Addr)
	}

	ctx := context.Background()
	err = serve(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	config.PressEnterToExit()
}
