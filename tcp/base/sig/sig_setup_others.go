// Copyright © 2019 Hedzr Yeh.

// +build nacl plan9

package sig

import (
	"log"
	"os"
)

// // SetupSignals initialize all signal handlers
// func SetupSignals() {
// 	setupSignals()
// }

func setupSignalsCommon() {
	//
}

func setupSignals() {
	// for i := 1; i < 34; i++ {
	// 	daemon.SetSigHandler(termHandler, syscall.Signal(i))
	// }

	// signals := []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGINT, syscall.SIGKILL, syscall.SIGUSR1, syscall.SIGUSR2}
	// if onSetTermHandler != nil {
	// 	signals = onSetTermHandler()
	// }
	// SetSigHandler(termHandler, signals...)
	//
	// signals = []os.Signal{syscall.Signal(0x7)}
	// if onSetSigEmtHandler != nil {
	// 	signals = onSetSigEmtHandler()
	// }
	// SetSigHandler(sigEmtHandler, signals...)
	//
	// signals = []os.Signal{syscall.SIGHUP}
	// if onSetReloadHandler != nil {
	// 	signals = onSetReloadHandler()
	// }
	// SetSigHandler(reloadHandler, signals...)
}

func termHandler(sig os.Signal) error {
	log.Println("terminating...")
	stop <- struct{}{}
	// log.Println("  - send stop ch")
	// if sig == syscall.SIGQUIT {
	// 	log.Println("  - waiting for done ch...")
	// 	<-done
	// 	log.Println("  - done ch received.")
	// }
	return ErrStop
}

func sigEmtHandler(sig os.Signal) error {
	log.Println("terminating (SIGEMT)...")
	stop <- struct{}{}
	// if sig == syscall.SIGQUIT {
	// 	<-done
	// }
	return ErrStop
}

func makeHandlers() (signals []os.Signal) {
	signals = make([]os.Signal, 0, len(handlers))
	for sig := range handlers {
		signals = append(signals, sig)
	}
	return
}

func nilSigSend(process *os.Process) error {
	// return process.Signal(syscall.Signal(0))
	return nil
}

func sigSendHUP(process *os.Process) error {
	// return process.Signal(syscall.SIGHUP)
	return nil
}

func sigSendUSR1(process *os.Process) error {
	return nil
}

func sigSendUSR2(process *os.Process) error {
	return nil
}

func sigSendTERM(process *os.Process) error {
	// return process.Signal(syscall.SIGTERM)
	return nil
}

func sigSendQUIT(process *os.Process) error {
	return nil // process.Signal(syscall.SIGQUIT)
}

func sigSendKILL(process *os.Process) error {
	return nil // process.Signal(syscall.SIGKILL)
}

// QuitSignal return a channel for quit signal raising up.
func QuitSignal() chan os.Signal {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	return quit
}
