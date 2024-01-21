package main

import (
	"context"
	stdnet "net"
	"os"
	"sync"

	logz "log/slog"

	"github.com/hedzr/is"

	"github.com/hedzr/go-socketlib/_examples"
	"github.com/hedzr/go-socketlib/net"
)

func init() {
	// println("OK")
	// logz.SetLevel(logz.DebugLevel)
	// logz.AddFlags(logz.Lprivacypath | logz.Lprivacypathregexp)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// logger := logz.New("new-dns")
	// logger := logz.Default()
	// logger := logz.New(logz.NewTextHandler(os.Stderr, &logz.HandlerOptions{Level: logz.LevelDebug}))
	logger := logz.New(_examples.NewPrettyHandler(os.Stdout, _examples.PrettyHandlerOptions{
		SlogOpts: logz.HandlerOptions{Level: logz.LevelDebug, AddSource: true},
	}))

	server := net.NewServer(":7099",
		net.WithServerOnListening(func(ss net.Server, l stdnet.Listener) {
			go runClient(ctx, ss, l, logger)
		}),
		// net.WithServerLogger(logger.WithSkip(1)),
		net.WithServerLogger(logger),
	)
	defer server.Close()

	catcher := is.Signals().Catch()
	catcher.
		// WithVerboseFn(func(msg string, args ...any) {
		// 	logz.WithSkip(2).Verbose(fmt.Sprintf("[verbose] %s", fmt.Sprintf(msg, args...)))
		// }).
		WithOnSignalCaught(func(sig os.Signal, wg *sync.WaitGroup) {
			println()
			logz.Debug("signal caught", "sig", sig)
			if err := server.Shutdown(); err != nil {
				logz.Error("server shutdown error", "err", err)
			}
			cancel()
		}).
		Wait(func(stopChan chan<- os.Signal, wgShutdown *sync.WaitGroup) {
			logz.Debug("entering looper's loop...")

			server.WithOnShutdown(func(err error, ss net.Server) { wgShutdown.Done() })
			err := server.ListenAndServe(ctx, nil)
			if err != nil {
				server.Fatal("server serve failed", "err", err)
			}
		})
}

func runClient(ctx context.Context, ss net.Server, l stdnet.Listener, logger net.Logger) {
	c := net.NewClient(net.WithClientLogger(logger))

	if err := c.Dial("tcp", ":7099"); err != nil {
		c.Fatal("connecting to server failed", "err", err, "server-endpoint", ":7099")
	}
	c.Info("[client] connected", "server.addr", c.RemoteAddr())
	c.RunDemo(ctx)
}
