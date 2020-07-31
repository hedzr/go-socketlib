package msg

import (
	"fmt"
	"gopkg.in/hedzr/errors.v2"
)

type Type uint8
type TKL uint8
type OptionNumber uint32

type Message struct {
	Type      Type
	TKL       TKL
	Code      uint8
	MessageID uint16
	Token     uint64
	Options   []Opt
	Payload   Payload
}

func (s *Message) Encode() (output []byte, err error) {
	return
}

func (s *Message) Decode(data []byte) (err error) {
	var i, newI int
	var b byte = data[i]
	ver := b >> 6
	if ver != 1 {
		err = errors.New("invalid VER: %v", ver)
		return
	}
	s.Type = Type(b>>4) & 3
	s.TKL = TKL(b & 0x0f)

	i++
	s.Code = data[i]
	s.MessageID = uint16(data[i+1])*256 + uint16(data[i+2]) // Big-endian
	for i = 4; i < 4+int(s.TKL); i++ {
		s.Token = s.Token<<8 + uint64(data[i])
	}

	// Options if any
	if newI, err = s.decodeOptions(data, i); err != nil {
		return
	}

	// Payload if any
	s.Payload = payloadCreate(data, newI)
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
	var optType int

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
		if length != 15 {
			err = errors.New("wrong: expect payload lead byte but got: Option delta (%v) and length (%v)", delta, length)
			return
		}

		// found payload leading byte, return and do payload extracting
		newIndex--
		return
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

	optType += delta
	s.Options = append(s.Options, newAnyOpt(optType, data[newIndex:newIndex+length]))
	newIndex += length
	goto nextOption
}

func (s *Message) String() string {
	//var sb strings.Builder
	//sb.WriteRune('[')
	//for i,opt:=range s.Options
	//sb.WriteRune(']')

	return fmt.Sprintf("Type: %v, TKL: %v, Code: %v, MsgID: %04x/%d, Token: %x. Options: %v. Payload: %v",
		s.Type, s.TKL, s.Code, s.MessageID, s.MessageID, s.Token,
		s.Options,
		s.Payload)
}

func (s *Message) As() {

}

func (s *Message) AsBytes() (out []byte) {
	return
}
