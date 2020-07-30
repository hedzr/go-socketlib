package client

import (
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/base"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tcp/udp"
	"net"
	"os"
	"sync"
	"time"
)

func tcpUnixBenchLoop(config *base.Config, done chan bool, opts ...Opt) (err error) {
	maxTimes := cmdr.GetIntRP(config.PrefixInCommandLine, "times")
	parallel := cmdr.GetIntRP(config.PrefixInCommandLine, "parallel")
	sleep := cmdr.GetDurationRP(config.PrefixInCommandLine, "sleep")

	var wg sync.WaitGroup
	wg.Add(parallel)
	for x := 0; x < parallel; x++ {
		go clientRunner(config.Logger, config.PrefixInCommandLine, x, config.Addr, maxTimes, sleep, &wg)
	}
	wg.Wait()

	done <- true // to end the TrapSignalsEnh waiter by manually, instead of os signals caught.
	return
}

func tcpUnixLoop(config *base.Config, opts ...Opt) (err error) {
	var conn net.Conn
	var done = make(chan bool, 1)
	var tid = 1

	ctc := tls2.NewCmdrTlsConfig(config.PrefixInConfigFile, config.PrefixInCommandLine)
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
	co.startLoopers()

	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		config.Logger.Debugf("signal[%v] caught and exiting this program", s)
	})()
	return
}

func udpLoop(config *base.Config, opts ...Opt) (err error) {
	ctx := context.Background()
	defer func() {
		config.Logger.Debugf("end.")
	}()

	co := newClientObj(nil, config.Logger, opts...)
	defer co.Close()

	uo := udp.NewUdpObj(co, nil, nil)
	if err = uo.Connect(ctx, config.Network, config); err != nil {
		config.Logger.Errorf("failed to create udp socket handler: %v", err)
	}
	go func() {
		if err = uo.Serve(ctx); err != nil {
			config.Logger.Errorf("failed to communicate via udp socket handler: %v", err)
		}
		config.Logger.Debugf("Serve() end.")
	}()

	_, err = uo.RawWrite(ctx, []byte("hello"))
	//uo.WriteTo(nil, []byte("hello"))
	config.Logger.Debugf("'hello' wrote: %v", err)

	//_, err = uo.WriteThrough([]byte("world"))
	uo.WriteTo(nil, []byte("world"))
	config.Logger.Debugf("'world' wrote: %v", err)

	time.Sleep(time.Second)
	config.PressEnterToExit()
	// _, _ = uo.WriteThrough([]byte("hello"))

	//n, data := 0, make([]byte, 1024)
	//n, err = conn.Read(data)
	//fmt.Printf("read %s from <%s>\n", data[:n], conn.RemoteAddr())

	return
}
