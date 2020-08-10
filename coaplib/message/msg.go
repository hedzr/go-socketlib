package message

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"gopkg.in/hedzr/errors.v2"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"
)

type Message struct {
	Type      Type
	TKL       TKL
	Code      Code
	MessageID uint16
	Token     uint64
	Options   []Opt
	Payload   Payload
	err       error
	ts        time.Time

	OnACK   OnACKHandler
	OnEvent OnEventHandler

	// available in received message

	MediaType MediaType
}

type OnACKHandler func(ctx context.Context, sent, recv *Message) (err error)
type OnEventHandler func(ctx context.Context, key string, token uint64, recv *Message)

func (s *Message) FindOption(num OptionNumber) (opt Opt) {
	opt = FindOption(num, s.Options)
	return
}

func (s *Message) FindOptions(num OptionNumber) (opt []Opt) {
	opt = FindOptions(num, s.Options)
	return
}

func (s *Message) CalcDigitLen(i uint64) (length int) {
	switch {
	case i > 0xffffffffffffff:
		length = 8
	case i > 0xffffffffffff:
		length = 7
	case i > 0xffffffffff:
		length = 6
	case i > 0xffffffff:
		length = 5
	case i > 0xffffff:
		length = 4
	case i > 0xffff:
		length = 3
	case i > 0xff:
		length = 2
	case i == 0:
		length = 0
	default:
		length = 1
	}
	return
}

func (s *Message) WithOnACK(fn OnACKHandler) {
	s.OnACK = fn
}

func (s *Message) WithOnEvent(fn OnEventHandler) {
	s.OnEvent = fn
}

func (s *Message) RaiseOnEvent(ctx context.Context, key string, token uint64, recv *Message) {
	if s.OnEvent != nil {
		s.OnEvent(ctx, key, token, recv)
	}
}

func (s *Message) SetToken(token uint64) {
	if token != 0 {
		s.Token = token
		s.TKL = TKL(s.CalcDigitLen(token))
	}
}

func (s *Message) Encode() (output []byte, err error) {
	output, err = s.AsBytes(), s.err
	return
}

func (s *Message) Decode(data []byte) (err error) {
	var i, newI int
	var b = data[i]

	ver := b >> 6
	if ver != 1 {
		err = errors.New("invalid VER: %v", ver)
		return
	}

	s.TKL = TKL(b & 0x0f)
	if s.TKL >= 9 {
		err = errors.New("invalid TKL: %v", s.TKL)
		return
	}

	s.ts, s.MediaType, s.Type = time.Now().UTC(), MediaTypeUndefined, Type(b>>4)&3

	i++
	s.Code = Code(data[i])
	if strings.HasPrefix(s.Code.String(), "Code(") {
		err = errors.New("invalid Code: %v(%s)", int(s.Code), s.Code)
		return
	}

	s.MessageID = uint16(data[i+1])*256 + uint16(data[i+2]) // Big-endian

	for i = 4; i < 4+int(s.TKL); i++ {
		s.Token = s.Token<<8 + uint64(data[i]) // Big-endian
	}

	// Options if any
	if newI, err = s.decodeOptions(data, i); err != nil {
		return
	}

	if opt := s.FindOption(OptionNumberContentFormat); opt != nil {
		s.MediaType = opt.(*optMediaType).MediaType
	}

	// Payload if any
	s.Payload = payloadCreate(data, newI, s.MediaType, s.Options)
	return
}

// https://tools.ietf.org/html/rfc5198
func (s *Message) decodeString(data []byte) (res string) {
	res = string(data)
	return
}

// https://tools.ietf.org/html/rfc5198
func (s *Message) encodeString(data string) (res []byte) {
	res = []byte(data)
	return
}

func (s *Message) decodeOptions(data []byte, i int) (newIndex int, err error) {
	var optNum OptionNumber

	newIndex = i

nextOption:
	if newIndex >= len(data) {
		return
	}

	delta, length := int(data[newIndex]>>4), int(data[newIndex]&0x0f)
	newIndex++

	switch delta {
	case 13:
		delta += int(data[newIndex])
		newIndex++
	case 14:
		delta += int(data[newIndex])*256 + int(data[newIndex+1])
		newIndex += 2
	case 15:
		if length == 15 {
			//err = errors.New("wrong: expect payload lead byte but got: Option delta (%v) and length (%v)", delta, length)
			//return
			//}

			// found payload leading byte, return and do payload extracting
			newIndex--
			return
		}
	default:
	}

	switch length {
	case 13:
		length += int(data[newIndex])
		newIndex++
	case 14:
		length += int(data[newIndex])*256 + int(data[newIndex+1])
		newIndex += 2
	case 15:
		if length != 15 {
			err = errors.New("wrong: expect payload lead byte but got: Option delta (%v) and length (%v)", delta, length)
			return
		}
	default:
	}

	optNum += OptionNumber(delta)
	if opt := optDecoder.Decode(optNum, data[newIndex:newIndex+length]); opt != nil {
		s.Options = append(s.Options, opt)
	} else {
		logger.Errorf("option %q [% x] decoding failed: %v", optNum, data[newIndex:newIndex+length], optDecoder.Error())
	}
	newIndex += length
	goto nextOption
}

func (s *Message) String() string {
	//var sb strings.Builder
	//sb.WriteRune('[')
	//for i,opt:=range s.Options
	//sb.WriteRune(']')

	l := 0
	if s.Payload != nil {
		l = s.Payload.Size()
	}

	return fmt.Sprintf("Type: %v, TKL: %v, Code: %v, MsgID: %04x/%d, Token: %x. Options: %v. Payload (%d bytes): |%s|",
		s.Type, s.TKL, s.Code, s.MessageID, s.MessageID, s.Token,
		s.Options,
		l, s.Payload)
}

func (s *Message) Href() string {
	var uri url.URL
	opts := s.FindOptions(OptionNumberURIHost)
	for _, opt := range opts {
		if o, ok := opt.(interface{ StringData() string }); ok {
			uri.Host = o.StringData()
		}
	}
	opts = s.FindOptions(OptionNumberURIPath)
	for _, opt := range opts {
		if o, ok := opt.(interface{ StringData() string }); ok {
			uri.Path = path.Join(uri.Path, o.StringData())
		}
	}
	opts = s.FindOptions(OptionNumberURIQuery)
	var q []string
	for _, opt := range opts {
		if o, ok := opt.(interface{ StringData() string }); ok {
			q = append(q, o.StringData())
		}
	}
	uri.RawQuery = strings.Join(q, "&")

	href := uri.Path
	if len(href) > 0 && href[0] != '/' {
		href = "/" + href
	}
	return href
}

func (s *Message) As() {

}

func (s *Message) AsBytes() (out []byte) {
	var bb bytes.Buffer

	s.err = bb.WriteByte(verBits | byte(s.Type<<4) | byte(s.TKL))
	if s.err != nil {
		return
	}

	s.err = bb.WriteByte(byte(s.Code))
	if s.err != nil {
		return
	}

	s.err = binary.Write(&bb, binary.BigEndian, s.MessageID)
	if s.err != nil {
		return
	}

	s.err = binary.Write(&bb, binary.BigEndian, s.Token)
	if s.err != nil {
		return
	}

	var (
		//optNum     = OptionNumberReserved
		lastOptNum    = OptionNumberReserved
		length        int
		delta         OptionNumber
		data          []byte
		oneByteBuffer = []byte{0}
	)

	tSlice := make(sort.IntSlice, len(s.Options))
	for i, o := range s.Options {
		tSlice[i] = (int(o.Number()) << 16) + i
	}
	sort.Slice(tSlice, tSlice.Less)

	//lastON,length := OptionNumberReserved,0
	for _, ix := range tSlice {
		o := s.Options[ix&0xffff]
		if o != nil {
			delta, data = o.Number()-lastOptNum, o.Bytes()
			lastOptNum, length = o.Number(), len(data)

			s.err = o.WriteOptDeltaAndLength(&bb, oneByteBuffer, delta, length)
			if s.err != nil {
				return
			}
			if length > 0 {
				_, s.err = bb.Write(data)
				if s.err != nil {
					return
				}
			}
		}
	}

	if s.Payload != nil {
		s.err = bb.WriteByte(0xff)
		_, s.err = bb.Write(s.Payload.Bytes())
	}

	// s.err = w.Flush()
	out = bb.Bytes()
	s.ts = time.Now().UTC()
	return
}

const ver = byte(1)
const verBits = ver << 6
