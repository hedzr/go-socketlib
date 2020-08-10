package pi

import (
	"bytes"
	"context"
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"github.com/hedzr/go-socketlib/tcp/base"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

func NewCoapCmd() *CoapCmd {
	return &CoapCmd{
		builder:          newBuilder(),
		ackBuilder:       newBuilder(),
		sentMessages:     map[uint16]*message.Message{},
		wellKnownCoreRes: message.NewLinkFormat(),
		observed:         map[uint64]*message.Message{}, // key: Token
		maxRegistered:    1,
	}
}

type CoapCmd struct {
	builder    *Builder
	ackBuilder *Builder
	done       chan bool
	config     *base.Config
	conn       base.Conn

	tryDebugMode bool
	dryRunMode   bool

	sentMessages map[uint16]*message.Message

	wellKnownCoreBuffer bytes.Buffer
	wellKnownCoreRes    *message.LinkFormat

	observed           map[uint64]*message.Message
	inRegisteringAll   int32
	inUnregisteringAll int32
	unregisteredCh     chan struct{}
	maxRegistered      int
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

	if cc.tryDebugMode {
		cc.Discover(ctx)

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
	}

	time.Sleep(time.Second)
	config.PressEnterToExit()

	cc.unregisteredCh = make(chan struct{})
	_ = cc.doUnregisterAll(ctx)
	<-cc.unregisteredCh
	cc.unregisteredCh = nil
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

func (cc *CoapCmd) Discover(ctx context.Context) {
	cc.Get(ctx, wellKnownCoreUri, WithOpOpOnAck(cc.OnAckWellKnownCore))
}

func (cc *CoapCmd) ACK(ctx context.Context, mid uint16) (err error) {
	const typ = message.ACK
	const cod = message.CodeEmpty

	msg := cc.ackBuilder.Reset().WithType(typ).WithCode(cod).WithMessageID(mid).Build()
	err = cc.write(ctx, msg)
	return
}

func (cc *CoapCmd) Get(ctx context.Context, uriPath string, opts ...OperateOpt) {
	ooo := NewOperateOpt(opts...)
	// .With(WithOpOpMsgCode(message.MethodCodePOST), WithOpOpContentFormat(message.TextPlain))
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) Post(ctx context.Context, uriPath string, opts ...OperateOpt) {
	ooo := NewOperateOpt(opts...).With(WithOpOpMsgCode(message.MethodCodePOST), WithOpOpContentFormat(message.TextPlain))
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) Put(ctx context.Context, uriPath string, opts ...OperateOpt) {
	ooo := NewOperateOpt(opts...).With(WithOpOpMsgCode(message.MethodCodePUT), WithOpOpContentFormat(message.TextPlain))
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) Delete(ctx context.Context, uriPath string, opts ...OperateOpt) {
	ooo := NewOperateOpt(opts...).With(WithOpOpMsgCode(message.MethodCodeDELETE))
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) Register(ctx context.Context, uriPath string, opts ...OperateOpt) {
	ooo := NewOperateOpt(opts...).
		With(
			WithOpOpRegister(0),
			WithOpOpOnAck(cc.OnRegistered),
			WithOpOpOnEvent(cc.OnEvent),
		)
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) Unregister(ctx context.Context, uriPath string, opts ...OperateOpt) {
	//var token uint64
	//for tok, sent := range cc.observed {
	//	if sent.Href() == uriPath {
	//		token = tok
	//		break
	//	}
	//}

	ooo := NewOperateOpt(opts...).With(
		WithOpOpRegister(1),
		//WithOpOpObserverToken(token),
		WithOpOpOnAck(cc.OnRegistered),
	)
	cc.doOperate(ctx, uriPath, ooo)
}

func (cc *CoapCmd) doOperate(ctx context.Context, uriPath string, ooo *opOp) {

	var (
		msg *message.Message
		err error
	)

	cc.builder. // Reset().WithURIString(uri)
			NewBase(cc.config.UriBase).
			WithURIPath(uriPath). // ("/.well-known/core") // weird33 // test
			WithOnACK(ooo.onAck).
			WithOnEvent(ooo.onEvent).
			WithType(ooo.typ).
			WithCode(ooo.cod).
			WithRegister(ooo.reg).
			WithMessageOptions(ooo.opts...).
			WithToken(ooo.token)

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

func (cc *CoapCmd) doObserveAll(ctx context.Context, sent, recv *message.Message) (err error) {
	atomic.AddInt32(&cc.inRegisteringAll, 1)
	err = cc.doObserveAllImpl(ctx)
	return
}

func (cc *CoapCmd) doObserveAllImpl(ctx context.Context) (err error) {
	if len(cc.observed) < cc.maxRegistered {
		for path, obs := range cc.wellKnownCoreRes.Observables {
			if _, ok := cc.observed[obs.ObserverToken]; !ok {
				cc.conn.Logger().Infof("  -> observing %q, %q", path, obs.Title)
				var opts []OperateOpt
				if obs.ContentType != message.MediaTypeUndefined {
					opts = append(opts, WithOpOpAccept(obs.ContentType))
				}
				opts = append(opts, WithOpOpObserverToken(obs.ObserverToken))
				cc.Register(ctx, path, opts...)
				return
			}
		}
	}
	atomic.AddInt32(&cc.inRegisteringAll, -1)
	return
}

func (cc *CoapCmd) doUnregisterAll(ctx context.Context) (err error) {
	atomic.AddInt32(&cc.inUnregisteringAll, 1)
	err = cc.doUnregisterAllImpl(ctx)
	return
}

func (cc *CoapCmd) doUnregisterAllImpl(ctx context.Context) (err error) {
	for token, sent := range cc.observed {
		path := sent.Href()
		cc.conn.Logger().Infof("  -> unregistering the observed %q [tok: %X]", path, token)
		var opts []OperateOpt
		opts = append(opts, WithOpOpObserverToken(token))
		cc.Unregister(ctx, path, opts...)
		return
	}

	atomic.AddInt32(&cc.inUnregisteringAll, -1)
	close(cc.unregisteredCh)
	return
}

func (cc *CoapCmd) OnEvent(ctx context.Context, key string, token uint64, recv *message.Message) {
	var seq uint64
	if opt := recv.FindOption(message.OptionNumberObserve); opt != nil {
		if o, ok := opt.(interface{ Uint64Data() uint64 }); ok {
			seq = o.Uint64Data()
		}
	}
	cc.conn.Logger().Infof("<-- observing %q: #%v [%X]: %v", key, seq, token, recv.Payload)
}

func (cc *CoapCmd) OnRegistered(ctx context.Context, sent, recv *message.Message) (err error) {
	if opt := sent.FindOption(message.OptionNumberObserve); opt != nil {
		if o, ok := opt.(interface{ Uint64Data() uint64 }); ok {
			switch o.Uint64Data() {
			case 0: // register
				key := sent.Token // sent.Href()
				cc.observed[key] = sent
				sent.RaiseOnEvent(ctx, sent.Href(), key, recv)
				err = cc.doObserveAllImpl(ctx)
			case 1: // deregister
				delete(cc.observed, recv.Token)
				err = cc.doUnregisterAllImpl(ctx)
			}
		}
	}
	return
}

func (cc *CoapCmd) OnAckWellKnownCore(ctx context.Context, sent, recv *message.Message) (err error) {
	err = cc.OnAckLinkFormat(ctx, sent, recv)

	opt := message.FindOption(message.OptionNumberBlock2, recv.Options)
	if optBlock2, ok := opt.(*message.OptBlockN); ok {
		if !optBlock2.More {

			// decoding the whole link format string

			err = cc.wellKnownCoreRes.Parse(cc.wellKnownCoreBuffer.String())
			if err != nil {
				cc.conn.Logger().Errorf("[COAP] combine all well-known/core packages failed: %v", err)
				return
			}

			cc.conn.Logger().Debugf("[COAP] well-known/core packages parsed OK: %v", cc.wellKnownCoreRes.String())

			if cc.tryDebugMode {

				// loop for each resource declared in well-known lists

				for path, res := range cc.wellKnownCoreRes.Resources {
					if !strings.EqualFold(path, wellKnownCoreUri) {
						cc.conn.Logger().Infof("  -> querying %q, %q", path, res.Title)
					}
				}

				err = cc.doObserveAll(ctx, sent, recv)
			}
		}
	}
	return
}

func (cc *CoapCmd) OnAckLinkFormat(ctx context.Context, sent, recv *message.Message) (err error) {
	// cc.conn.Logger().Debugf("   ⦿ sent, recv := %v, %v", sent, recv)

	opt := message.FindOption(message.OptionNumberBlock2, recv.Options)
	if optBlock2, ok := opt.(*message.OptBlockN); ok {

		if recv.Payload != nil {
			cc.wellKnownCoreBuffer.Write(recv.Payload.Bytes())
		}

		if optBlock2.More {

			// retrieve the next block

			cc.builder.From(sent, removeBlock2OptAlways).
				WithRequestBlock1Num(int(optBlock2.Num+1), uint(optBlock2.SizeInBytes()))

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
			return
		}

	}

	return
}

const (
	removeBlock2OptAlways = true

	wellKnownCoreUri = "/.well-known/core"
)
