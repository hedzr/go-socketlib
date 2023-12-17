module github.com/hedzr/go-socketlib

go 1.17

//replace github.com/hedzr/log => ../../go-cmdr/10.log

//replace github.com/hedzr/logex => ../../go-cmdr/15.logex

//replace github.com/hedzr/cmdr => ../../go-cmdr/50.cmdr

// replace github.com/hedzr/go-ringbuf => ../go-ringbuf

// replace github.com/hedzr/rules => ../rules

// replace github.com/hedzr/pools => ../pools

// replace github.com/hedzr/errors => ../errors

require (
	github.com/hedzr/cmdr v1.11.25
	github.com/hedzr/log v1.6.25
	github.com/hedzr/logex v1.6.25
	gopkg.in/hedzr/errors.v3 v3.3.0
	gopkg.in/hedzr/go-ringbuf.v1 v1.0.3
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/hedzr/cmdr-base v1.0.0 // indirect
	github.com/hedzr/evendeep v0.4.17 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/crypto v0.16.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
