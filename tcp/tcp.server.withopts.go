/*
 * Copyright © 2020 Hedzr Yeh.
 */

package tcp

import (
	log2 "log"

	"github.com/hedzr/go-socketlib/tcp/tls"
	"github.com/hedzr/log"
)

func WithServerOnProcessFunc(onProcess OnTcpServerProcessFunc) ServerOpt {
	return func(server *Server) {
		server.onTcpProcess = onProcess
	}
}

func WithServerBufferSize(size int) ServerOpt {
	return func(server *Server) {
		server.bufferSize = size
		if size <= 0 {
			log2.Panicf("wrong buffer size: %v", size)
		}
	}
}

func WithServerReadWriter(onCreateReadWriter OnTcpServerCreateReadWriter) ServerOpt {
	return func(server *Server) {
		server.onTcpServerCreateReadWriter = onCreateReadWriter
	}
}

func WithServerDisconnectedWithClient(fn OnTcpServerDisconnectedWithClient) ServerOpt {
	return func(server *Server) {
		server.onTcpServerDisconnectedWithClient = fn
	}
}

func WithServerConnectedWithClient(fn OnTcpServerConnectedWithClient) ServerOpt {
	return func(server *Server) {
		server.onTcpServerConnectedWithClient = fn
	}
}

func WithServerListening(fn OnTcpServerListening) ServerOpt {
	return func(server *Server) {
		server.onTcpServerListening = fn
	}
}

func WithTlsConfig(s *tls.CmdrTlsConfig) ServerOpt {
	return func(server *Server) {
		server.CmdrTlsConfig = s
	}
}

//func WithLoggerConfig(config *log.LoggerConfig) ServerOpt {
//	return func(server *Server) {
//		server.Logger = build.New(config)
//	}
//}

func WithLogger(l log.Logger) ServerOpt {
	return func(server *Server) {
		server.Logger = l
	}
}
