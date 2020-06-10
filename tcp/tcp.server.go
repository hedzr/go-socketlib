/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	"github.com/hedzr/cmdr"
	"github.com/sirupsen/logrus"
	"os"
)

type ServerOpt func(*Server)
type ClientOpt func(*Client)

func StartServer(addr string, opts ...ServerOpt) *Server {
	s := newServer(addr, opts...)
	if err := s.Start(); err != nil {
		s.Wrong(err, "can't start tcp server (addr=%v)", addr)
	}
	return s
}

func StopServer(s *Server) {
	s.Stop()
}

func HandleSignals(onTrapped func(s os.Signal)) (waiter func()) {
	waiter = cmdr.TrapSignals(onTrapped)
	return
}

func model1() {
	doneChan := make(chan interface{})

	go func(done <-chan interface{}) {
		defer func() {
			logrus.Debug("child goroutine exited.")
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
