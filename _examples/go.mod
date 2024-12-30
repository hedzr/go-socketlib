module github.com/hedzr/go-socketlib/_examples

go 1.22.7

//replace github.com/hedzr/is => ../../../cmdr.v2/libs.is

//replace github.com/hedzr/env => ../../../cmdr.v2/libs.env

//replace github.com/hedzr/logg => ../../../cmdr.v2/libs.logg

replace github.com/hedzr/go-socketlib => ../

require (
	github.com/hedzr/go-socketlib v1.1.1
	github.com/hedzr/is v0.6.5
	github.com/miekg/dns v1.1.62
)

require (
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/term v0.27.0 // indirect
	golang.org/x/tools v0.28.0 // indirect
)
