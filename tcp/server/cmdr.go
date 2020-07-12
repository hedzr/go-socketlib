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
		Description("TCP Server Operations").
		Group("Test").
		Action(serverRun)

	tcpServer.NewFlagV(b.port, "port", "p").
		Description("The port to listen on").
		Group("Test").
		Placeholder("PORT")

	tcpServer.NewFlagV("", "addr", "a", "adr", "address").
		Description("The address to listen to").
		Group("Test").
		Placeholder("HOST-or-IP")

	tcpServer.NewFlagV(false, "enable-tls", "tls").
		Description("enable TLS mode").
		Group("TLS")
	
	tcpServer.NewFlagV("root.pem", "cacert", "ca", "ca-cert").
		Description("CA cert path (.cer,.crt,.pem) if it's standalone").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.pem", "cert", "c").
		Description("server public-cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("cert.key", "key", "k").
		Description("server private-key path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV(false, "client-auth").
		Description("enable client cert authentication").
		Group("TLS")
	tcpServer.NewFlagV(2, "tls-version").
		Description("tls-version: 0,1,2,3").
		Group("TLS")

}
