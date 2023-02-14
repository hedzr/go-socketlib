package opts

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/tcp/cert"
	"github.com/hedzr/go-socketlib/tcp/client"
	"github.com/hedzr/go-socketlib/tcp/server"
)

func AttachToCmdr(root cmdr.OptCmd) {
	socketLibCmd(root)
}

func socketLibCmd(root cmdr.OptCmd) {

	// TCP

	tcpCmd := root.NewSubCommand("tcp", "t", "socketlib").
		Description("go-socketlib TCO operations...", "").
		Group("Socket")

	server.AttachToCmdrCommand(tcpCmd,
		server.WithCmdrPort(1983),
		server.WithCmdrServerProtocolInterceptor(newServerPI()),
	)
	client.AttachToCmdrCommand(tcpCmd,
		client.WithCmdrPort(1983),
		client.WithCmdrInteractiveCommand(true),
		client.WithCmdrClientProtocolInterceptor(newClientPI()),
		client.WithCmdrClientBuildPackageFunc(buildPkg),
	)

	// UDP

	udpCmd := root.NewSubCommand("udp", "u").
		Description("go-socketlib UDP operations...", "").
		Group("Socket")

	server.AttachToCmdrCommand(udpCmd,
		server.WithCmdrUDPMode(true),
		server.WithCmdrPort(1984),
	)
	client.AttachToCmdrCommand(udpCmd,
		client.WithCmdrUDPMode(true),
		client.WithCmdrPort(1984),
	)

	// Cert

	cert.AttachToCmdrCommand(root)

}
