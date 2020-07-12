/*
 * Copyright © 2020 Hedzr Yeh.
 */

package client

import (
	"github.com/hedzr/cmdr"
	"time"
)

const (
	DefaultPort = 8883
)

func AttachToCmdr(tcp cmdr.OptCmd) {
	// tcp := root.NewSubCommand().
	// 	Titles("t", "tcp").
	// 	Description("", "").
	// 	Group("Test")
	// // Action(func(cmd *cmdr.Command, args []string) (err error) {
	// // 	return
	// // })

	tcpClient := tcp.NewSubCommand("client", "c").
		Description("TCP client operations").
		Group("Test").
		Action(run)

	tcpClient.NewFlagV(DefaultPort, "port", "p").
		Description("The port to connect to").
		Group("Test").
		Placeholder("PORT")

	tcpClient.NewFlagV("127.0.0.1", "host", "h", "address", "addr").
		Description("The hostname or IP to connect to").
		Group("Test").
		Placeholder("HOST-or-IP")
	// don't use localhost, it may cause 'lookup localhost: no such host' error in debug mode.

	tcpClient.NewFlagV(100, "times", "t").
		Description("repeat sending times").
		Group("Test").
		Placeholder("n")

	tcpClient.NewFlagV(3, "parallel", "r").
		Description("how many clients parallel").
		Group("Test").
		Placeholder("n")

	tcpClient.NewFlagV(time.Duration(0), "sleep").
		Description("sleep time between each sending").
		Group("Test")

	tcpClient.NewFlagV(false, "interactive", "i").
		Description("run client in interactive mode").
		Group("Test")

	tcpClient.NewFlagV("", "cafile", "ca").
		Description("CA cert path (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpClient.NewFlagV("", "cert").
		Description("client public-cert path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpClient.NewFlagV("", "key").
		Description("client private-key path for dual auth (.cer,.crt,.pem)").
		Group("TLS").
		Placeholder("PATH")
	tcpClient.NewFlagV(false, "client-auth").
		Description("enable client cert authentication").
		Group("TLS")
	tcpClient.NewFlagV(2, "tls-version").
		Description("tls-version: 0,1,2,3").
		Group("TLS")

}
