module github.com/hedzr/socketlib

go 1.13

replace github.com/hedzr/ringbuf => ../go-ringbuf

// replace github.com/hedzr/rules => ../rules

// replace github.com/hedzr/pools => ../pools

// replace github.com/hedzr/errors => ../errors

require (
	github.com/hedzr/cmdr v1.6.47
	github.com/hedzr/errors v1.1.18
	github.com/hedzr/logex v1.1.8
	github.com/hedzr/ringbuf v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.6.0
	gitlab.com/hedzr/mqttlib v1.0.5
)
