/*
 * Copyright © 2020 Hedzr Yeh.
 */

package client

import (
	"context"
	"fmt"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/base"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tool"
	"github.com/hedzr/log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

//const prefixSuffix = "client.tls"
const defaultNetType = "tcp"

type CommandAction func(cmd *cmdr.Command, args []string, mainLoop MainLoop, prefixPrefix string, opts ...Opt) (err error)
type MainLoop func(ctx context.Context, conn base.Conn, done chan bool, config *base.Config)
type MainLoopHolder interface {
	MainLoop(ctx context.Context, conn base.Conn, done chan bool, config *base.Config)
}

func DefaultLooper(cmd *cmdr.Command, args []string, mainLoop MainLoop, prefixPrefix string, opts ...Opt) (err error) {
	config := base.NewConfigFromCmdrCommand(false, prefixPrefix, cmd)
	config.BuildLogger()
	if err = config.BuildAddr(); err != nil {
		config.Logger.Fatalf("%v", err)
	}

	if strings.HasPrefix(config.Network, "udp") {
		err = udpLoop(config, mainLoop, opts...)
		return
	}

	err = tcpUnixLoop(config, mainLoop, opts...)
	return
}

func runAsCliTool(cmd *cmdr.Command, args []string, mainLoop MainLoop, prefixPrefix string, opts ...Opt) (err error) {
	config := base.NewConfigFromCmdrCommand(false, prefixPrefix, cmd)
	config.BuildLogger()
	if err = config.BuildAddr(); err != nil {
		config.Logger.Fatalf("%v", err)
	}

	if strings.HasPrefix(config.Network, "udp") {
		err = udpLoop(config, mainLoop, opts...)
		return
	}

	// tcp, unix

	done := make(chan bool, 1)
	if cmdr.GetBool("interactive", cmdr.GetBoolRP(config.PrefixInCommandLine, "interactive")) {
		err = runOneClient(config.Logger, done, cmd, args)

	} else {
		err = tcpUnixBenchLoop(config, done, opts...)

	}

	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		config.Logger.Debugf("signal[%v] caught and exiting this program", s)
	})()
	return
}

func interactiveRunAsCliTool(cmd *cmdr.Command, args []string) (err error) {
	return
}

func clientRunner(logger log.Logger, prefixCLI string, tid int, dest string, maxTimes int, sleep time.Duration, wg *sync.WaitGroup) {
	var (
		err  error
		conn net.Conn
	)
	defer wg.Done()

	netType := cmdr.GetStringRP(prefixCLI, "network", "tcp")
	prefix := netType + ".client.tls"
	// prefixCLI := "tcp.client"

	ctc := tls2.NewCmdrTlsConfig(prefix, prefixCLI)
	logger.Debugf("%v", ctc)
	logger.Debugf("dest: %v", dest)

	// netType := cmdr.GetStringRP(prefix, "network", netType)
	conn, err = ctc.Dial(netType, dest)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			// fmt.Println("Some problem connecting.")
			logger.Errorf("[%d] Some problem connecting: %v", tid, err)
		} else {
			// fmt.Println("Unknown error: " + err.Error())
			logger.Errorf("[%d] failed: %v", tid, err)
		}
		// os.Exit(1)
		return
	}

	co := newClientObj(conn, logger)
	go co.readConnection()

	for i := 0; i < maxTimes; i++ {
		// text := fmt.Sprintf("%d.%d. %v", tid, i, randRandSeq())
		err = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			// fmt.Println("Error set writing deadline.")
			logger.Errorf("[%d] failed to write daedline: %v", tid, err)
			break
		}
		// _, err = conn.Write([]byte(text))
		_, err = conn.Write(mqttConnectPkg)
		if err != nil {
			// fmt.Println("Error writing to stream.")
			logger.Errorf("[%d] failed to write to stream: %v", tid, err)
			break
		}
		logger.Debugf(" #%d sent", i)
		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
	return
}

func runOneClient(logger log.Logger, done chan bool, cmd *cmdr.Command, args []string) (err error) {
	prefixInCommandLine := cmd.GetDottedNamePath()
	dest := fmt.Sprintf("%s:%v", cmdr.GetStringRP(prefixInCommandLine, "host"), cmdr.GetIntRP(prefixInCommandLine, "port"))
	fmt.Printf("Connecting to %s...\n", dest)

	var conn net.Conn
	//conn, err = net.Dial("tcp", dest)

	prefix := "tcp.client.tls"
	// prefixCLI := "tcp.client"
	ctc := tls2.NewCmdrTlsConfig(prefix, prefixInCommandLine)
	logger.Debugf("%v", ctc)
	logger.Debugf("dest: %v", dest)
	conn, err = ctc.Dial("tcp", dest)

	if err != nil {
		if _, t := err.(*net.OpError); t {
			logger.Errorf("Some problem connecting. error: %v", err)
		} else {
			logger.Errorf("Unknown error: %v", err)
		}
		done <- true // to end the TrapSignalsEnh waiter by manually, instead of os signals caught.
		os.Exit(1)
	}

	newClientObj(conn, logger).run()

	done <- true // to end the TrapSignalsEnh waiter by manually, instead of os signals caught.
	return
}

// https://github.com/aaronbieber/tcp-server-client-go/blob/master/client/client.go
// var host = flag.String("host", "localhost", "The hostname or IP to connect to; defaults to \"localhost\".")
// var port = flag.Int("port", 8000, "The port to connect to; defaults to 8000.")
//
// func main() {
// 	flag.Parse()
//
// 	dest := *host + ":" + strconv.Itoa(*port)
// 	fmt.Printf("Connecting to %s...\n", dest)
//
// 	conn, err := net.Dial("tcp", dest)
//
// 	if err != nil {
// 		if _, t := err.(*net.OpError); t {
// 			fmt.Println("Some problem connecting.")
// 		} else {
// 			fmt.Println("Unknown error: " + err.Error())
// 		}
// 		os.Exit(1)
// 	}
//
// 	go readConnection(conn)
//
// 	for {
// 		reader := bufio.NewReader(os.Stdin)
// 		fmt.Print("> ")
// 		text, _ := reader.ReadString('\n')
//
// 		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
// 		_, err := conn.Write([]byte(text))
// 		if err != nil {
// 			fmt.Println("Error writing to stream.")
// 			break
// 		}
// 	}
// }

func randRandSeq() string {
	n := rand.Intn(64)
	return tool.RandSeqln(n)
}

var mqttConnectPkg = []byte{
	16, 35,
	0, 4, 77, 81, 84, 84, 4, 2, 0, 60,
	0, 23, 109, 111, 115, 113, 45, 52, 100, 72, 113,
	105, 78, 86, 99, 88, 110, 50, 65, 69, 77, 103,
	99, 90, 78,
}

var mqttSubsribePkg = []byte{
	130, 12,
	0, 1,
	0, 7, 116, 111, 112, 105, 99, 48, 49, 0,
}

var mqttSubscribe2TopicsPkg = []byte{
	130, 22,
	0, 2,
	0, 7, 116, 111, 112, 105, 99, 48, 49, 0,
	0, 7, 116, 111, 112, 105, 99, 48, 50, 0,
}
