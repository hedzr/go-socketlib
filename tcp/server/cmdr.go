/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"github.com/hedzr/cmdr"
)

const (
	DefaultPort = 8883
)

type Opt func(*builder)

type builder struct {
	port int
}

func WithPort(port int) Opt {
	return func(b *builder) {
		b.port = port
	}
}

func AttachToCmdr(tcp cmdr.OptCmd, opts ...Opt) {
	// tcp := root.NewSubCommand().
	// 	Titles("t", "tcp").
	// 	Description("", "").
	// 	Group("Test")
	// // Action(func(cmd *cmdr.Command, args []string) (err error) {
	// // 	return
	// // })

	b := &builder{
		port: DefaultPort,
	}
	for _, opt := range opts {
		opt(b)
	}

	tcpServer := tcp.NewSubCommand("server", "s").
		Description("TCP/UDP Server Operations").
		Group("Test").
		Action(serverRun)

	cmdr.NewBool().Titles("stop", "s", "shutdown").
		Description("stop/shutdown the running server").
		Group("Tool").
		AttachTo(tcpServer)

	tcpServer.NewFlagV(b.port, "port", "p").
		Description("The port to listen on").
		Group("TCP/UDP").
		Placeholder("PORT")

	tcpServer.NewFlagV("", "addr", "a", "adr", "address").
		Description("The address to listen to").
		Group("TCP/UDP").
		Placeholder("HOST-or-IP")

	cmdr.NewBool().Titles("0001.enable-tls", "tls").
		Description("enable TLS mode").
		Group("TLS").
		AttachTo(tcpServer)
	//tcpServer.NewFlagV(false, "enable-tls", "tls").
	//	Description("enable TLS mode").
	//	Group("TLS")

	tcpServer.NewFlagV("root.pem", "100.cacert", "ca", "ca-cert").
		Description("CA cert path (.cer,.crt,.pem) if it's standalone").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.pem", "110.cert", "c").
		Description("server public-cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.key", "120.key", "k").
		Description("server private-key path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV(false, "190.client-auth").
		Description("enable client cert authentication").
		Group("TLS")
	tcpServer.NewFlagV(2, "200.tls-version").
		Description("tls-version: 0,1,2,3").
		Group("TLS")

}
