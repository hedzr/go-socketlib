/*
 * Copyright © 2020 Hedzr Yeh.
 */

package server

import (
	"github.com/hedzr/cmdr"
	mqtool "gitlab.com/hedzr/mqttlib"
)

func AttachToCmdr(tcp cmdr.OptCmd) {
	// tcp := root.NewSubCommand().
	// 	Titles("t", "tcp").
	// 	Description("", "").
	// 	Group("Test")
	// // Action(func(cmd *cmdr.Command, args []string) (err error) {
	// // 	return
	// // })

	tcpServer := tcp.NewSubCommand("server", "s").
		Description("TCP Server Operations").
		Group("Test").
		Action(serverRun)

	tcpServer.NewFlagV(mqtool.DefaultPort, "port", "p").
		Description("The port to listen on").
		Group("Test").
		Placeholder("PORT")

	tcpServer.NewFlagV("", "addr", "a", "adr", "address").
		Description("The address to listen to").
		Group("Test").
		Placeholder("HOST-or-IP")

	tcpServer.NewFlagV("", "cafile", "ca").
		Description("CA cert path (.cer,.crt,.pem) if it's standalone").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("").
		Titles("", "cert").
		Description("server public-cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpServer.NewFlagV("", "key").
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
