/*
 * Copyright © 2019 Hedzr Yeh.
 */

package sig

import (
	"log"
	"net"
	"os"
	"os/signal"

	"gopkg.in/hedzr/errors.v3"
)

// ServeSignals calls handlers for system signals.
// before invoking ServeSignals(), you should run SetupSignals() at first.
func ServeSignals() (err error) {
	if len(handlers) == 0 {
		setupSignals()
	}
	if len(handlers) == 0 {
		return // no handlers, skip the os signal listening
	}

	// syscall.Getenv()
	signals := makeHandlers()

	// defer func() {
	// 	removePID(ctx)
	// }()
	//
	// log.Printf("serve signals ... pid: %v in %v", os.Getpid(), ctx.PidFileName)

	ch := make(chan os.Signal, 8)
	signal.Notify(ch, signals...)

	for sig := range ch {
		log.Printf(".. signal caught: %v", sig)
		err = handlers[sig](sig)
		if err != nil {
			break
		}
	}

	signal.Stop(ch)

	if err == ErrStop {
		err = nil
	}

	return
}

// HandleSignalCaughtEvent is a shortcut to block the main business logic loop but break it if os signals caught.
// `stop` channel will be trigger if any hooked os signal caught, such as os.Interrupt;
// the main business logic loop should trigger `done` once `stop` holds.
func HandleSignalCaughtEvent() bool {
	select {
	case <-stop:
		log.Print("stop ch received. send done ch.")
		done <- struct{}{}
		return true
	default:
		return false
	}
}

// GetChs returns the standard `stop`, `done` channel
func GetChs() (stopCh, doneCh chan struct{}) {
	stopCh, doneCh = stop, done
	return
}

var (
	// ErrStop should be returned signal handler function
	// for termination of handling signals.
	ErrStop = errors.New("stop serve signals")

	handlers = make(map[os.Signal]SignalHandlerFunc)

	// child *os.Process

	onSetTermHandler      func() []os.Signal
	onSetSigEmtHandler    func() []os.Signal
	onSetReloadHandler    func() []os.Signal
	onSetHotReloadHandler func() []os.Signal
	onGetListener         func() net.Listener

	stop = make(chan struct{})
	done = make(chan struct{})
)

const (
	// ErrnoForkAndDaemonFailed is os errno when daemon plugin and its impl occurs errors.
	ErrnoForkAndDaemonFailed = -1
	envvarInDaemonized       = "__DAEMON"
)

// SignalHandlerFunc is the interface for signal handler functions.
type SignalHandlerFunc func(sig os.Signal) (err error)
