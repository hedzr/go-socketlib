package pi

import (
	"bytes"
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"github.com/hedzr/go-socketlib/tcp/base"
	"net"
	"time"
)

func NewCoapCmd() *CoapCmd {
	return &CoapCmd{
		builder:          NewBuilder(),
		sentMessages:     map[uint16]*message.Message{},
		wellKnownCoreRes: message.NewLinkFormat(),
	}
}

type CoapCmd struct {
	builder      *Builder
	done         chan bool
	config       *base.Config
	conn         base.Conn
	tryDebugMode bool
	dryRunMode   bool

	sentMessages map[uint16]*message.Message

	wellKnownCoreBuffer bytes.Buffer
	wellKnownCoreRes    *message.LinkFormat
}

func (cc *CoapCmd) MainLoop(ctx context.Context, conn base.Conn, done chan bool, config *base.Config) {
	//if wr, ok := conn.(base.CachedUDPWriter); ok {
	//	// coap://coap.me/404
	//	data, err := hex.DecodeString("480109e4b0fb52f90c76043f37636f61702e6d658334303460")
	//	if err == nil {
	//		wr.WriteTo(nil, data)
	//	}
	//}

	cc.conn, cc.config, cc.done = conn, config, done
	cc.tryDebugMode = cmdr.GetBoolRP(config.PrefixInCommandLine, "try-debug")
	cc.dryRunMode = cmdr.GetBoolRP(config.PrefixInCommandLine, "dry-run")

	// connected

	cc.Get(ctx, "/.well-known/core")

	//var (
	//	msg *message.Message
	//	err error
	//)
	//
	//builder := NewBuilder()
	//builder.
	//	NewBase("coap://coap.me").
	//	WithURIPath("/.well-known/core") // weird33 // test
	//msg = builder.Build()
	//if builder.Error() != nil {
	//	cc.conn.Logger().Fatalf("can't build coap diagram: %v", builder.Error())
	//}
	//
	//if cc.tryDebugMode {
	//	return
	//}
	//
	//if !cc.dryRunMode {
	//	if err = cc.write(ctx, msg); err != nil {
	//		cc.conn.Logger().Errorf("write failed: %v", err)
	//	}
	//}

	time.Sleep(time.Second)
	config.PressEnterToExit()
}

func (cc *CoapCmd) SetBase(remoteAddr net.Addr) {
}

func (cc *CoapCmd) write(ctx context.Context, msg *message.Message) (err error) {
	if c, ok := cc.conn.(base.CachedUDPWriter); ok {
		c.WriteTo(nil, msg.AsBytes())
	} else {
		_, err = cc.conn.RawWrite(ctx, msg.AsBytes())
	}
	return
}

func (cc *CoapCmd) Get(ctx context.Context, uriPath string) {

	var (
		msg *message.Message
		err error
	)

	cc.builder. // Reset().WithURIString(uri)
			NewBase(cc.config.UriBase).
			WithURIPath(uriPath). // ("/.well-known/core") // weird33 // test
			WithOnACK(cc.onAckWellKnownCore)

	msg = cc.builder.Build()
	if cc.builder.Error() != nil {
		cc.conn.Logger().Fatalf("can't build coap diagram: %v", cc.builder.Error())
	}

	if err = cc.write(ctx, msg); err != nil {
		cc.conn.Logger().Errorf("write failed: %v", err)
	} else {
		cc.sentMessages[msg.MessageID] = msg
		cc.conn.Logger().Debugf("[COAP] ⬆ Get: %v", msg)
	}
}

func (cc *CoapCmd) onAckWellKnownCore(ctx context.Context, sent, recv *message.Message) (err error) {
	cc.conn.Logger().Debugf("   ⦿ sent, recv := %v, %v", sent, recv)

	opt := message.FindOption(message.OptionNumberBlock2, recv.Options)
	if optBlock2, ok := opt.(*message.OptBlockN); ok {

		if recv.Payload != nil {
			cc.wellKnownCoreBuffer.Write(recv.Payload.Bytes())
		}

		if optBlock2.More {
			cc.builder.From(sent, removeBlock2OptAlways).
				WithRequestBlock2Num(int(optBlock2.Num+1), uint(optBlock2.SizeInBytes()))

			msg := cc.builder.Build()
			if cc.builder.Error() != nil {
				cc.conn.Logger().Fatalf("can't build coap diagram: %v", cc.builder.Error())
			}

			if err = cc.write(ctx, msg); err != nil {
				cc.conn.Logger().Errorf("write failed: %v", err)
			} else {
				cc.sentMessages[msg.MessageID] = msg
				cc.conn.Logger().Debugf("[COAP] ⬆ Get: %v", msg)
			}
		} else {
			err = cc.wellKnownCoreRes.Parse(cc.wellKnownCoreBuffer.String())
			if err == nil {
				cc.conn.Logger().Debugf("[COAP] well-known/core packages parsed OK: %v", cc.wellKnownCoreRes.String())
			} else {
				cc.conn.Logger().Errorf("[COAP] combine all well-known/core packages failed: %v", err)
			}
		}
	}

	return
}

const removeBlock2OptAlways = true
