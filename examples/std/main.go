package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hedzr/go-socketlib/tcp/base"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/log"
	"os"
	"time"
)

var (
	clientMode     = flag.Bool("client", false, "server (false) or client (true)")
	host           = flag.String("host", "localhost", "listening host")
	port           = flag.Int("port", 50001, "listening port")
	reg            = flag.String("reg", "localhost:32379", "register etcd address")
	count          = flag.Int("count", 3, "instance's count")
	connectTimeout = flag.Duration("connect-timeout", 5*time.Second, "connect timeout")
)

func main() {

	flag.Parse()

	if *clientMode == false {
		runServer()
	} else {
		runClient()
	}
}

func runClient() {
	_ = os.Setenv("APPNAME", "std-client")

	// default is zap sugar logger
	logConfig := log.NewLoggerConfig()
	logConfig.Backend = "logrus" // zap, sugar, std, off/dummy/none
	logConfig.Level = "debug"

	var ignoredKey, ignoredAdapterName, ignoredUriBase string
	config := base.NewConfigWithParams(false,
		"tcp",      // tcp, udp, or unix
		ignoredKey, // ignore safely because you give up from cmdr
		ignoredKey, // ignore safely because you give up from cmdr
		logConfig,
		func(cfg *tls2.CmdrTlsConfig) {
			cfg.Cert = "unknown" // give a valid path if you like to enable TLS
			cfg.Key = "unknown"
		},
		"localhost:1983",
		ignoredUriBase,
		ignoredAdapterName,
	)

	err := client.New(false, config,
		client.WithClientMainLoop(func(ctx context.Context, conn base.Conn, done chan bool, config *base.Config) {
			_, _ = conn.RawWrite(ctx, []byte(fmt.Sprintf("hello %v", time.Now())))
			config.PressEnterToExit()
		}),
	)
	if err != nil {
		panic(err)
	}
}

func runServer() {
	// app-name, these component need it: pid-file, log-file, ...
	_ = os.Setenv("APPNAME", "std-server")

	// default is zap sugar logger
	logConfig := log.NewLoggerConfig()
	logConfig.Backend = "logrus" // zap, sugar, std, off/dummy/none
	logConfig.Level = "debug"

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

	serve, serverObj, tlsEnabled, err := server.New(config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if tlsEnabled {
		log.Printf("Listening on %s with TLS enabled.", config.Addr)
	} else {
		log.Printf("Listening on %s.", config.Addr)
	}

	go func() {
		ctx := context.Background()
		err = serve(ctx)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}()

	config.PressEnterToExit()
	server.Shutdown(serverObj)
}
