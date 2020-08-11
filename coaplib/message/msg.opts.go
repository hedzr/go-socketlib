package message

import (
	"math/rand"
	"time"
)

// New create an empty Message for the further operations
func New(code Code, opts ...MsgOpt) *Message {
	m := &Message{
		Type:      CON,
		TKL:       0,
		Code:      code,
		MessageID: 0,
		Token:     rand.Uint64(),
		Options:   nil,
		Payload:   nil,
		err:       nil,
		ts:        time.Now().UTC(),
		OnACK:     nil,
		OnEvent:   nil,
		MediaType: 0,
	}

	for _, opt := range opts {
		opt(m)
	}
	return m
}

// FindOption lookups and finds out an option within a Message Option(option) array
func FindOption(num OptionNumber, options []Option) (opt Option) {
	for _, o := range options {
		if o.Number() == num {
			opt = o
			break
		}
	}
	return
}

// FindOption lookups and finds out all matched options within a Message Option(option) array
func FindOptions(num OptionNumber, options []Option) (opt []Option) {
	for _, o := range options {
		if o.Number() == num {
			opt = append(opt, o)
		}
	}
	return
}

// MsgOpt is the option of creating a new Message via New().
type MsgOpt func(s *Message)

// WithType:
//     Type: CON, NON, ACK, RST
func WithType(typ Type) MsgOpt {
	return func(s *Message) {
		s.Type = typ
	}
}

func WithMessageID(mid uint32) MsgOpt {
	return func(s *Message) {
		s.MessageID = uint16(mid)
	}
}

func WithToken(token uint64) MsgOpt {
	return func(s *Message) {
		s.SetToken(token)
	}
}

func WithOnACK(fn OnACKHandler) MsgOpt {
	return func(s *Message) {
		s.WithOnACK(fn)
	}
}

func WithOnEvent(fn OnEventHandler) MsgOpt {
	return func(s *Message) {
		s.WithOnEvent(fn)
	}
}
