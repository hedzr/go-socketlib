package pi

import (
	"context"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"math/rand"
)

type Operable interface {
	Get(ctx context.Context, uriPath string, opts ...OperateOpt)
	Put(ctx context.Context, uriPath string, opts ...OperateOpt)
	Post(ctx context.Context, uriPath string, opts ...OperateOpt)
	Delete(ctx context.Context, uriPath string, opts ...OperateOpt)
	Register(ctx context.Context, uriPath string, opts ...OperateOpt)
	Unregister(ctx context.Context, uriPath string, opts ...OperateOpt)
}

type Discoverable interface {
	Discover(ctx context.Context)
}

type Serveable interface {
	ACK(ctx context.Context, mid uint16) (err error)
}

type OperateOpt func(*opOp)

type opOp struct {
	typ      message.Type
	cod      message.Code
	onAck    message.OnACKHandler
	onEvent  message.OnEventHandler
	opts     []message.Opt
	postData []byte
	reg      int
	token    uint64
}

func (op *opOp) With(opts ...OperateOpt) *opOp {
	for _, o := range opts {
		o(op)
	}
	return op
}

func NewOperateOpt(opts ...OperateOpt) *opOp {
	oo := &opOp{typ: message.CON, cod: message.MethodCodeGET, reg: -1}
	for _, o := range opts {
		if o != nil {
			o(oo)
		}
	}
	return oo
}

func WithOpOpOnAck(handler message.OnACKHandler) OperateOpt {
	return func(op *opOp) {
		op.onAck = handler
	}
}

func WithOpOpOnEvent(handler message.OnEventHandler) OperateOpt {
	return func(op *opOp) {
		op.onEvent = handler
	}
}

func WithOpOpMsgType(typ message.Type) OperateOpt {
	return func(op *opOp) {
		op.typ = typ
	}
}

func WithOpOpMsgCode(cod message.Code) OperateOpt {
	return func(op *opOp) {
		op.cod = cod
	}
}

func WithOpOpRegister(reg int) OperateOpt {
	return func(op *opOp) {
		op.reg = reg
		if op.token == 0 {
			op.token = rand.Uint64()
		}
	}
}

func WithOpOpObserverToken(token uint64) OperateOpt {
	return func(op *opOp) {
		op.token = token
	}
}

func WithOpOpPostData(data []byte) OperateOpt {
	return func(op *opOp) {
		op.postData = data
		WithOpOpMsgCode(message.MethodCodePOST)(op)
	}
}

func WithOpOpContentFormat(mt message.MediaType) OperateOpt {
	return func(op *opOp) {
		opt := message.NewUint64Opt(message.OptionNumberContentFormat, uint64(mt))
		op.opts = append(op.opts, opt)
	}
}

func WithOpOpAccept(accept message.MediaType) OperateOpt {
	return func(op *opOp) {
		opt := message.NewUint64Opt(message.OptionNumberAccept, uint64(accept))
		op.opts = append(op.opts, opt)
	}
}
