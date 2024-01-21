package main

import (
	"context"
	"fmt"
	stdnet "net"
	"os"
	"sync"

	logz "log/slog"

	"github.com/hedzr/is"

	"github.com/hedzr/go-socketlib/_examples"
	"github.com/hedzr/go-socketlib/net"
	"github.com/hedzr/go-socketlib/net/api"
)

func init() {
	// println("OK")
	// logz.SetLevel(logz.DebugLevel)
	// logz.AddFlags(logz.Lprivacypath | logz.Lprivacypathregexp)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// logger := logz.New("new-udp")
	// logger := logz.Default()
	// logger := logz.New(logz.NewTextHandler(os.Stderr, &logz.HandlerOptions{Level: logz.LevelDebug}))
	logger := logz.New(_examples.NewPrettyHandler(os.Stdout, _examples.PrettyHandlerOptions{
		SlogOpts: logz.HandlerOptions{Level: logz.LevelDebug, AddSource: true},
	}))

	server := net.NewServer(":7099",
		net.WithNetwork("udp"),
		net.WithServerOnListening(func(ss net.Server, l stdnet.Listener) {
			go runClient(ctx, ss, l, logger)
		}),
		// net.WithServerLogger(logger.WithSkip(1)),
		net.WithServerLogger(logger),
		net.WithServerOnProcessData(func(data []byte, w api.Response, r api.Request) (nn int, err error) {
			logz.Debug("[server] RECEIVED:", "data", string(data), "client.addr", w.RemoteAddr())
			nn = len(data)
			w.Write(data)
			return
		}),
	)
	defer server.Close()

	is.Signals().Catch().
		WithVerboseFn(func(msg string, args ...any) {
			// logger.WithSkip(2).Verbose(fmt.Sprintf("[verbose] %s", fmt.Sprintf(msg, args...)))
			server.Verbose(fmt.Sprintf("[verbose] %s", fmt.Sprintf(msg, args...)))
		}).
		WithOnSignalCaught(func(sig os.Signal, wg *sync.WaitGroup) {
			println()
			logger.Debug("signal caught", "sig", sig)
			if err := server.Shutdown(); err != nil {
				logger.Error("server shutdown error", "err", err)
			}
			cancel()
		}).
		Wait(func(stopChan chan<- os.Signal, wgShutdown *sync.WaitGroup) {
			// server.Debug("entering looper's loop...")

			go func() {
				server.WithOnShutdown(func(err error, ss net.Server) { wgShutdown.Done() })
				err := server.ListenAndServe(ctx, nil)
				if err != nil {
					server.Fatal("server serve failed", "err", err)
				}
			}()
		})
}

func runClient(ctx context.Context, ss net.Server, l stdnet.Listener, logger net.Logger) {
	c := net.NewClient(net.WithClientLogger(logger))

	if err := c.Dial("udp", ":7099"); err != nil {
		c.Fatal("connecting to server failed", "err", err, "server-endpoint", ":7099")
	}
	c.Info("[client] connected", "server.addr", c.RemoteAddr())
	c.RunDemo(ctx)
}
