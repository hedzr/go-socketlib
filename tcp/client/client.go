/*
 * Copyright © 2020 Hedzr Yeh.
 */

package client

import (
	"fmt"
	"github.com/hedzr/cmdr"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/go-socketlib/tool"
	"github.com/hedzr/logex"
	"github.com/hedzr/logex/build"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

func run(cmd *cmdr.Command, args []string) (err error) {
	loggerConfig := build.NewLoggerConfig()
	_ = cmdr.GetSectionFrom("logger", &loggerConfig)
	logger := build.New(loggerConfig)

	done := make(chan bool, 1)
	prefixInCommandLine := cmd.GetDottedNamePath()

	if cmdr.GetBool("interactive", cmdr.GetBoolRP(prefixInCommandLine, "interactive")) {
		err = runOneClient(logger, cmd, args)

	} else {
		dest := fmt.Sprintf("%s:%v", cmdr.GetStringRP(prefixInCommandLine, "host"), cmdr.GetIntRP(prefixInCommandLine, "port"))
		maxTimes := cmdr.GetIntRP(prefixInCommandLine, "times")
		parallel := cmdr.GetIntRP(prefixInCommandLine, "parallel")
		sleep := cmdr.GetDurationRP(prefixInCommandLine, "sleep")
		var wg sync.WaitGroup
		wg.Add(parallel)
		for x := 0; x < parallel; x++ {
			go clientRunner(logger, prefixInCommandLine, x, dest, maxTimes, sleep, &wg)
		}
		wg.Wait()

		done <- true // to end the TrapSignalsEnh waiter by manually, instead of os signals caught.
	}

	cmdr.TrapSignalsEnh(done, func(s os.Signal) {
		logger.Debugf("signal[%v] caught and exiting this program", s)
	})()
	return
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

func clientRun(cmd *cmdr.Command, args []string) (err error) {
	return
}

func randRandSeq() string {
	n := rand.Intn(64)
	return tool.RandSeqln(n)
}

func clientRunner(logger logex.Logger, prefixCLI string, tid int, dest string, maxTimes int, sleep time.Duration, wg *sync.WaitGroup) {
	var (
		err  error
		conn net.Conn
	)
	defer wg.Done()

	prefix := "tcp.client.tls"
	// prefixCLI := "tcp.client"
	ctc := tls2.NewCmdrTlsConfig(prefix, prefixCLI)
	logger.Debugf("%v", ctc)
	logger.Debugf("dest: %v", dest)
	conn, err = ctc.Dial("tcp", dest)

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

func runOneClient(logger logex.Logger, cmd *cmdr.Command, args []string) (err error) {
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
		os.Exit(1)
	}

	newClientObj(conn, logger).run()

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
