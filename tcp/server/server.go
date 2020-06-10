/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"bufio"
	"fmt"
	"github.com/hedzr/cmdr"
	tls2 "github.com/hedzr/socketlib/tcp/tls"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"strconv"
	"time"
)

//
// 1. 跨平台通用，而非性能优先，即并非针对具体平台优化，但在跨平台基础上优化性能。
// 2. 在 1 的基础上，讲究
//    1. 易于拓展：良好的插件机制
//    2. 易于使用：良好的API
// 3. 工业级稳定性、正确性。采用严格的测试，压测。
// 4. 性能优化，算法级别的优化。
// 5. 讲求代码美观
//    1. 不过渡设计，不绕圈子
//    2. 适当冗余以优化性能
// 6. IPv4, IPv6, tcp, udp 全模态支持

func serverRun(cmd *cmdr.Command, args []string) (err error) {
	fmt.Println("Starting server...")

	prefix := "tcp.server"

	// src := fmt.Sprintf("%s:%v", cmdr.GetStringR("tcp.server.addr"), cmdr.GetIntR("tcp.server.port"))

	var addr, host, port string
	host, port, err = net.SplitHostPort(cmdr.GetStringRP(prefix, "addr"))
	if err != nil {
		logrus.Errorf("get broker address failed: %v", err)
		return
	}
	if port == "" {
		port = strconv.FormatInt(cmdr.GetInt64RP(prefix, "port"), 10)
	}
	addr = net.JoinHostPort(host, port)

	var listener net.Listener
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		logrus.Fatal(err)
	}

	ctcPrefix := "tcp.server.tls"
	ctc := tls2.NewCmdrTlsConfig(ctcPrefix, prefix)
	listener, err = ctc.NewTlsListener(listener)
	if err != nil {
		logrus.Fatal(err)
	}

	fmt.Printf("Listening on %s.\n", addr)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Some connection error: %s\n", err)
			continue
		}

		go handleConnection(conn)
	}
	return
}

// func newTlsConfig(cacertFile, certFile, keyFile string, clientAuth bool) (config *tls.Config, err error) {
// 	var cert tls.Certificate
// 	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
// 	if err != nil {
// 		err = errors.New("error parsing X509 certificate/key pair").Attach(err)
// 		return
// 	}
// 	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
// 	if err != nil {
// 		err = errors.New("error parsing certificate").Attach(err)
// 		return
// 	}
//
// 	// Create TLSConfig
// 	// We will determine the cipher suites that we prefer.
// 	config = &tls.Config{
// 		Certificates: []tls.Certificate{cert},
// 		MinVersion:   tls.VersionTLS12,
// 	}
//
// 	// Require client certificates as needed
// 	if clientAuth {
// 		config.ClientAuth = tls.RequireAndVerifyClientCert
// 	}
//
// 	// Add in CAs if applicable.
// 	if cacertFile != "" {
// 		rootPEM, err := ioutil.ReadFile(cacertFile)
// 		if err != nil || rootPEM == nil {
// 			return nil, err
// 		}
// 		pool := x509.NewCertPool()
// 		ok := pool.AppendCertsFromPEM([]byte(rootPEM))
// 		if !ok {
// 			err = errors.New("failed to parse root ca certificate")
// 		}
// 		config.ClientCAs = pool
// 	}
// 	return
// }

//
// var addr = flag.String("addr", "", "The address to listen to; default is \"\" (all interfaces).")
// var port = flag.Int("port", 8000, "The port to listen on; default is 8000.")
//
// func Run() {
// 	flag.Parse()
//
// 	fmt.Println("Starting server...")
//
// 	src := *addr + ":" + strconv.Itoa(*port)
// 	listener, _ := net.Listen("tcp", src)
// 	fmt.Printf("Listening on %s.\n", src)
//
// 	defer listener.Close()
//
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Printf("Some connection error: %s\n", err)
// 		}
//
// 		go handleConnection(conn)
// 	}
// }

func handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("Client connected from " + remoteAddr)

	scanner := bufio.NewScanner(conn)

	for {
		ok := scanner.Scan()

		if !ok {
			break
		}

		handleMessage(scanner.Text(), conn)
	}

	fmt.Println("Client at " + remoteAddr + " disconnected.")
}

func handleMessage(message string, conn net.Conn) {
	fmt.Println("> " + message)

	if len(message) > 0 && message[0] == '/' {
		switch {
		case message == "/time":
			resp := "It is " + time.Now().String() + "\n"
			fmt.Print("< " + resp)
			conn.Write([]byte(resp))

		case message == "/quit":
			fmt.Println("Quitting.")
			conn.Write([]byte("I'm shutting down now.\n"))
			fmt.Println("< " + "%quit%")
			conn.Write([]byte("%quit%\n"))
			os.Exit(0)

		default:
			conn.Write([]byte("Unrecognized command.\n"))
		}
	}
}
