/*
 * Copyright © 2019 Hedzr Yeh.
 */

package sig

import (
	"net"
	"os"

	"github.com/hedzr/cmdr"
)

// IsRunningInDemonizedMode returns true if you are running under demonized mode.
// false means that you're running in normal console/tty mode.
func IsRunningInDemonizedMode() bool {
	// return cmdr.GetBoolR(envvarInDaemonized)
	// return isDemonized()
	return cmdr.GetBoolR("server.start.in-daemon")
}

// SetTermSignals allows an functor to provide a list of Signals
func SetTermSignals(sig func() []os.Signal) {
	onSetTermHandler = sig
}

// SetSigEmtSignals allows an functor to provide a list of Signals
func SetSigEmtSignals(sig func() []os.Signal) {
	onSetSigEmtHandler = sig
}

// SetReloadSignals allows an functor to provide a list of Signals
func SetReloadSignals(sig func() []os.Signal) {
	onSetReloadHandler = sig
}

// SetHotReloadSignals allows an functor to provide a list of Signals
func SetHotReloadSignals(sig func() []os.Signal) {
	onSetHotReloadHandler = sig
}

// SetOnGetListener returns tcp/http listener for daemon hot-restarting
func SetOnGetListener(fn func() net.Listener) {
	onGetListener = fn
}

// SetSigHandler sets handler for the given signals.
// SIGTERM has the default handler, he returns ErrStop.
func SetSigHandler(handler SignalHandlerFunc, signals ...os.Signal) {
	for _, sig := range signals {
		handlers[sig] = handler
	}
}

// SendNilSig sends the POSIX NUL signal
func SendNilSig(process *os.Process) error {
	return nilSigSend(process)
}

// SendHUP sends the POSIX HUP signal
func SendHUP(process *os.Process) error {
	return sigSendHUP(process)
}

// SendUSR1 sends the POSIX USR1 signal
func SendUSR1(process *os.Process) error {
	return sigSendUSR1(process)
}

// SendUSR2 sends the POSIX USR2 signal
func SendUSR2(process *os.Process) error {
	return sigSendUSR2(process)
}

// SendTERM sends the POSIX TERM signal
func SendTERM(process *os.Process) error {
	return sigSendTERM(process)
}

// SendQUIT sends the POSIX QUIT signal
func SendQUIT(process *os.Process) error {
	return sigSendQUIT(process)
}

// SendKILL sends the POSIX KILL signal
func SendKILL(process *os.Process) error {
	return sigSendKILL(process)
}
