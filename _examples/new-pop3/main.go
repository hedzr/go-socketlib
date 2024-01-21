package main

import (
	"context"
	"fmt"
	stdnet "net"
	"os"
	"sync"

	logz "log/slog"

	"github.com/hedzr/is"
	"github.com/hedzr/is/basics"

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

	// logger := logz.New("new-pop3")
	// logger := logz.Default()
	// logger := logz.New(logz.NewTextHandler(os.Stderr, &logz.HandlerOptions{Level: logz.LevelDebug, AddSource: true}))
	logger := logz.New(_examples.NewPrettyHandler(os.Stdout, _examples.PrettyHandlerOptions{
		SlogOpts: logz.HandlerOptions{Level: logz.LevelDebug, AddSource: true},
	}))

	pop3server := newPop3Server(
		net.WithServerOnListening(func(ss net.Server, l stdnet.Listener) {
			go runClient(ctx, ss, l, logger)
		}),
		// net.WithServerLogger(logger.WithSkip(1)),
		net.WithServerLogger(logger),
	)
	defer pop3server.Close()

	var looperS []basics.OnLooper
	catcher := is.Signals().Catch()
	catcher.
		WithVerboseFn(func(msg string, args ...any) {
			// logger.WithSkip(2).Verbose(fmt.Sprintf("[verbose] %s", fmt.Sprintf(msg, args...)))
			pop3server.Verbose(fmt.Sprintf("[verbose] %s", fmt.Sprintf(msg, args...)))
		}).
		WithOnLoop(looperS...).
		WithOnSignalCaught(func(sig os.Signal, wg *sync.WaitGroup) {
			println()
			pop3server.Debug("signal caught", "sig", sig)
			if err := pop3server.Shutdown(); err != nil {
				pop3server.Error("server shutdown error", "err", err)
			}
			cancel() // trigger shutting down 'pop3server', see pop3server.ListenAndServe(ctx) below...
		}).
		Wait(func(stopChan chan<- os.Signal, wgDone *sync.WaitGroup) {
			// pop3server.Debug("entering looper's loop...")

			// setup handler: close catcher's waiting looper while 'pop3server' shut down
			pop3server.WithOnShutdown(func(err error, ss net.Server) { wgDone.Done() })

			go func() {
				err := pop3server.ListenAndServe(ctx, nil)
				if err != nil {
					pop3server.Error("server serve failed", "err", err)
					panic(err)
				}
			}()
		})
}

const pop3serverAddress = ":1110"
