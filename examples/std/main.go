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

	_ = os.Setenv("APPNAME", "std-server")

	lc := log.NewLoggerConfig()

	config := base.NewConfigWithParams(true, "tcp",
		"tcp",
		"",
		lc,
		func(config *tls2.CmdrTlsConfig) {
			config.Cert = "unknown"
		},
	)
	config.Addr = ":1983"
	config.UriBase = "my-protocol://localhost:1983"

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
