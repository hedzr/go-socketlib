package cert

import (
	"time"

	"github.com/hedzr/cmdr"
)

func AttachToCmdrCommand(optCmd cmdr.OptCmd) {
	certCmd := cmdr.NewSubCmd().Titles("cert", "").
		Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
		Group("Tool").AttachTo(optCmd)

	certSubCommands(certCmd)
}

func certSubCommands(certOptCmd cmdr.OptCmd) {
	caCmd := cmdr.NewSubCmd().Titles("ca", "ca").
		Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
		Group("CA").
		AttachTo(certOptCmd)

	_ = cmdr.NewSubCmd().Titles("create", "c").
		Description("[NOT YET] create NEW CA certificates").
		Action(caCreate).
		AttachTo(caCmd)
	//log.Println(caCreateCmd)

	// certCmd := certOptCmd.NewSubCommand().
	// 	Titles("", "cert").
	// 	Description("certification tool (such as create-ca, create-cert, ...)", "certification tool (such as create-ca, create-cert, ...)\nverbose long descriptions here.").
	// 	Group("Tool")

	certCreateCmd := cmdr.NewSubCmd().Titles("create", "c").
		Description("create NEW certificates").
		Action(certCreate).
		AttachTo(certOptCmd)
	//log.Println(certCreateCmd)

	cmdr.NewStringSlice().Titles("hostnames", "hns").
		Description("Comma-separated hostname list and/or IPs to generate a certificate for").
		DefaultValue("localhost", "HOST").
		AttachTo(certCreateCmd)
	cmdr.NewString().Titles("start-date", "sd", "from", "valid-from").
		Description("Creation date formatted as Jan 1 15:04:05 2011 (default now)").
		AttachTo(certCreateCmd)
	cmdr.NewDuration(365*10*24*time.Hour).Titles("valid-for", "d", "duration").
		Description("Duration (10yr for debugging) that certificate is valid for").
		AttachTo(certCreateCmd)

	cmdr.NewString("./ci/certs").
		Titles("output-dir", "o", "output").
		Description("Creation date formatted as Jan 1 15:04:05 2011 (default now)").
		AttachTo(certCreateCmd)
}
