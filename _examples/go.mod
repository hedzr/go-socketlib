module github.com/hedzr/go-socketlib/_examples

go 1.23.0

toolchain go1.23.3

//replace github.com/hedzr/is => ../../../cmdr.v2/libs.is

//replace github.com/hedzr/env => ../../../cmdr.v2/libs.env

//replace github.com/hedzr/logg => ../../../cmdr.v2/libs.logg

replace github.com/hedzr/go-socketlib => ../

require (
	github.com/hedzr/go-socketlib v1.1.6
	github.com/hedzr/is v0.7.15
	github.com/miekg/dns v1.1.65
)

require (
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
)
