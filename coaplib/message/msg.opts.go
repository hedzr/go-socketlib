package message

import (
	"context"
	"math/rand"
)

func New(code Code, opts ...MsgOption) *Message {
	m := &Message{
		Type:      CON,
		TKL:       0,
		Code:      code,
		MessageID: 0,
		Token:     rand.Uint64(),
		Options:   nil,
		Payload:   nil,
		err:       nil,
	}

	for _, opt := range opts {
		opt(m)
	}
	return m
}

func FindOption(num OptionNumber, options []Opt) (opt Opt) {
	for _, o := range options {
		if o.Number() == num {
			opt = o
			break
		}
	}
	return
}

func FindOptions(num OptionNumber, options []Opt) (opt []Opt) {
	for _, o := range options {
		if o.Number() == num {
			opt = append(opt, o)
		}
	}
	return
}

type MsgOption func(s *Message)

// WithType:
//     Type: CON, NON, ACK, RST
func WithType(typ Type) MsgOption {
	return func(s *Message) {
		s.Type = typ
	}
}

func WithMessageID(mid uint32) MsgOption {
	return func(s *Message) {
		s.MessageID = uint16(mid)
	}
}

func WithToken(token uint64) MsgOption {
	return func(s *Message) {
		s.SetToken(token)
	}
}

func WithOnACK(fn func(ctx context.Context, sent, recv *Message) (err error)) MsgOption {
	return func(s *Message) {
		s.WithOnACK(fn)
	}
}
