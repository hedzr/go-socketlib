/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/cmdr"
	"log"
	"os"
)

type ServerOpt func(*Server)
type ClientOpt func(*Client)

func StartServer(addr string, opts ...ServerOpt) *Server {
	s := newServer(addr, opts...)
	if err := s.Start(); err != nil {
		s.Errorf("can't start tcp server (addr=%v): %v", addr, err)
	}
	return s
}

func StopServer(s *Server) {
	s.Stop()
}

// HandleSignals returns a waiter function to listen some predefined os signals.
//
//
// Usage
//
//  func enteringLoop() {
// 	  waiter := cmdr.HandleSignals(func(s os.Signal) {
// 	    // request shutdown:
//      globalExitChan <- true
// 	  })
// 	  go waiter()
//  }
//
//
//
func HandleSignals(onTrapped func(s os.Signal)) (waiter func()) {
	waiter = cmdr.TrapSignals(onTrapped)
	return
}

func model1() {
	doneChan := make(chan interface{})

	go func(done <-chan interface{}) {
		defer func() {
			log.Printf("child goroutine exited.")
		}()
		for {
			select {
			case <-done:
				return
			default:
			}
		}
	}(doneChan)

	close(doneChan)
}
