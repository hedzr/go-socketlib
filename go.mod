module github.com/hedzr/socketlib

go 1.13

// replace github.com/hedzr/rules v0.0.0 => ../rules

// replace github.com/hedzr/pools v0.0.0 => ../pools

// replace github.com/hedzr/errors v0.0.0 => ../errors

require (
	github.com/hedzr/cmdr v1.6.47
	github.com/hedzr/errors v1.1.18
	github.com/hedzr/logex v1.1.8
	github.com/sirupsen/logrus v1.6.0
	gitlab.com/hedzr/mqttlib v1.0.5
	go.uber.org/zap v1.10.0
	golang.org/x/sys v0.0.0-20200124204421-9fbb57f87de9
	gopkg.in/hedzr/errors.v2 v2.0.12
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
