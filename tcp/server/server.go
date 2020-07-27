/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"github.com/hedzr/cmdr"
	tls2 "github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/logex/build"
	"net"
	"strconv"
)

func serverRun(cmd *cmdr.Command, args []string) (err error) {

	loggerConfig := build.NewLoggerConfig()
	_ = cmdr.GetSectionFrom("logger", &loggerConfig)
	so := newServerObj(build.New(loggerConfig))

	so.logger.Infof("Starting server... cmdr.InDebugging = %v", cmdr.InDebugging())

	prefixInCommandLine := cmd.GetDottedNamePath()
	prefixInConfigFile := "tcp.server"

	var addr, host, port string
	host, port, err = net.SplitHostPort(cmdr.GetStringRP(prefixInConfigFile, "addr"))
	if port == "" {
		port = strconv.FormatInt(cmdr.GetInt64RP(prefixInConfigFile, "ports.default"), 10)
	}
	if port == "0" {
		port = strconv.FormatInt(cmdr.GetInt64RP(prefixInCommandLine, "port", 1024), 10)
		if port == "0" {
			so.logger.Fatalf("invalid port number: %q", port)
		}
	}
	addr = net.JoinHostPort(host, port)

	var listener net.Listener
	listener, err = serverBuildListener(so, addr, prefixInConfigFile, prefixInCommandLine)
	if err != nil {
		so.logger.Fatalf("build listener failed: %v", err)
	}
	so.Listen(listener)

	err = so.Serve()
	return
}

func serverBuildListener(so *Obj, addr, prefixInConfigFile, prefixInCommandLine string) (listener net.Listener, err error) {
	var tlsListener net.Listener
	listener, err = net.Listen(
		cmdr.GetStringRP(prefixInConfigFile, "network",
			cmdr.GetStringRP(prefixInCommandLine, "network", "tcp")),
		addr)
	if err != nil {
		so.logger.Fatalf("error: %v", err)
	}

	ctcPrefix := prefixInConfigFile + ".tls"
	ctc := tls2.NewCmdrTlsConfig(ctcPrefix, prefixInCommandLine)
	so.logger.Debugf("%v", ctc)
	if ctc.Enabled {
		tlsListener, err = ctc.NewTlsListener(listener)
		if err != nil {
			so.logger.Fatalf("error: %v", err)
		}
	}
	if tlsListener != nil {
		listener = tlsListener
		so.logger.Printf("Listening on %s with TLS enabled.", addr)
	} else {
		so.logger.Printf("Listening on %s.", addr)
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
