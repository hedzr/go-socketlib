package client

import (
	"context"
	"net"
	"sync"

	"github.com/hedzr/cmdr"
	"github.com/hedzr/log"

	"github.com/hedzr/go-socketlib/tcp/base"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tcp/udp"
)

func tcpUnixBenchLoop(config *base.Config, done chan bool, opts ...Opt) (err error) {
	maxTimes := cmdr.GetIntRP(config.PrefixInCommandLine, "times")
	parallel := cmdr.GetIntRP(config.PrefixInCommandLine, "parallel")
	sleep := cmdr.GetDurationRP(config.PrefixInCommandLine, "sleep")

	var wg sync.WaitGroup
	wg.Add(parallel)
	for x := 0; x < parallel; x++ {
		go clientRunner(config.Logger, config.PrefixInCommandLine, x, config.Addr, maxTimes, sleep, &wg, opts...)
	}
	wg.Wait()

	done <- true // to end the TrapSignalsEnh waiter by manually, instead of os signals caught.
	return
}

func tcpUnixLoop(config *base.Config, mainLoop MainLoop, opts ...Opt) (err error) {
	var (
		conn net.Conn
		done = make(chan bool, 1)
		tid  = 1
		ctc  *tls2.CmdrTlsConfig
	)
	if config.TlsConfigInitializer != nil {
		ctc = tls2.NewTlsConfig(config.TlsConfigInitializer)
	} else {
		ctcPrefix := config.PrefixInConfigFile + ".tls"
		ctc = tls2.NewCmdrTlsConfig(ctcPrefix, config.PrefixInCommandLine)
	}

	conn, err = ctc.Dial(config.Network, config.Addr)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			// fmt.Println("Some problem connecting.")
			config.Logger.Errorf("[%d] Some problem connecting: %v", tid, err)
		} else {
			// fmt.Println("Unknown error: " + err.Error())
			config.Logger.Errorf("[%d] failed: %v", tid, err)
		}
		// os.Exit(1)
		return
	}

	co := newClientObj(conn, config.Logger, opts...)
	defer co.Close()

	if i, ok := co.protocolInterceptor.(interface{ SetLogger(log.Logger) }); ok {
		i.SetLogger(config.Logger)
	}

	co.startLoopers(done)

	if mainLoop == nil {
		mainLoop = co.mainLoop
	}
	mainLoop(context.Background(), co.AsBaseConn(), done, config)
	// cmdr.TrapSignalsEnh(done, func(s os.Signal) {
	//	config.Logger.Debugf("signal[%v] caught and exiting this program", s)
	// })()
	return
}

func udpLoop(config *base.Config, mainLoop MainLoop, opts ...Opt) (err error) {
	ctx := context.Background()

	done := make(chan bool)
	defer func() {
		<-done
		config.Logger.Debugf("end.")
	}()

	co := newClientObj(nil, config.Logger, opts...)
	defer co.Join(ctx, done)

	if i, ok := co.protocolInterceptor.(interface{ SetLogger(log.Logger) }); ok {
		i.SetLogger(config.Logger)
	}

	ln := cmdr.GetIntRP(config.PrefixInConfigFile, "listeners", 0)
	uo := udp.New(co, udp.WithListenerNumber(ln))
	if err = uo.Connect(ctx, config); err != nil {
		config.Logger.Errorf("failed to create udp socket handler: %v", err)
		return
	}
	defer uo.Join(ctx, done)

	// co.SetBaseConn(uo.AsBaseConn())
	go func() {
		if err = uo.ClientServe(ctx); err != nil {
			config.Logger.Errorf("failed to communicate via udp socket handler: %v", err)
		}
		config.Logger.Debugf("Serve() end.")
	}()

	if mainLoop == nil {
		mainLoop = co.mainLoop
	}
	mainLoop(ctx, uo.AsBaseConn(), done, config)

	return
}
