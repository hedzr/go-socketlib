package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type optBase struct{}

func (s *optBase) Number() OptionNumber { return OptionNumberReserved }
func (s *optBase) String() string       { return "" }
func (s *optBase) Bytes() []byte        { return nil }
func (s *optBase) SetBytes(data []byte) {}

func (s *optBase) WriteOptDeltaAndLength(w io.Writer, oneByteBuffer []byte, delta OptionNumber, length int) (err error) {
	delta12, delta13, len12, len13 := int(delta), -1, length, -1
	if delta12 > 12 {
		if delta12 > 269 {
			delta13 = delta12 - 269
			delta12 = 14
		} else {
			delta13 = delta12 - 13
			delta12 = 13
		}
	}
	if len12 > 12 {
		if len12 > 269 {
			len13 = len12 - 269
			len12 = 14
		} else {
			len13 = len12 - 13
			len12 = 13
		}
	}

	oneByteBuffer[0] = byte(delta12<<4 | len12)
	_, err = w.Write(oneByteBuffer)
	if err != nil {
		return
	}

	if delta12 == 14 {
		oneByteBuffer[0] = byte(delta13 >> 8)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
		oneByteBuffer[0] = byte(delta13 & 0xff)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
	} else if delta12 == 13 {
		oneByteBuffer[0] = byte(delta13 & 0xff)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
	}
	if len12 == 14 {
		oneByteBuffer[0] = byte(len13 >> 8)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
		oneByteBuffer[0] = byte(len13 & 0xff)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
	} else if len12 == 13 {
		oneByteBuffer[0] = byte(len13 & 0xff)
		_, err = w.Write(oneByteBuffer)
		if err != nil {
			return
		}
	}
	return
}

//
//
//

func NewAnyOpt(optType OptionNumber, data []byte) *anyOpt {
	o := &anyOpt{
		Type: optType,
	}
	o.Data = make([]byte, len(data))
	copy(o.Data, data)
	return o
}

type anyOpt struct {
	optBase
	Type OptionNumber // opt number
	Data []byte       // include leading bytes: delta and length
}

func (s *anyOpt) Number() OptionNumber { return s.Type }
func (s *anyOpt) Bytes() []byte        { return s.Data }
func (s *anyOpt) SetBytes(data []byte) { s.Data = data }

func (s *anyOpt) String() string {
	return fmt.Sprintf("[%s: [% x]]", s.Type, s.Data)
}

//
//
//

func NewStringOpt(optType OptionNumber, data string) *stringOpt {
	o := &stringOpt{
		Type: optType,
		Data: data,
	}
	return o
}

type stringOpt struct {
	optBase
	Type OptionNumber
	Data string // without leading bytes
}

func (s *stringOpt) Number() OptionNumber { return s.Type }
func (s *stringOpt) Bytes() []byte        { return []byte(s.Data) }
func (s *stringOpt) StringData() string   { return s.Data }

func (s *stringOpt) String() string {
	return fmt.Sprintf("[%s: %q]", s.Type, s.Data)
}

//
//
//

func NewUint64Opt(optType OptionNumber, data uint64) Option {
	return &optUint64{
		Type: optType,
		Data: data,
	}
}

type optUint64 struct {
	optBase
	Type OptionNumber
	Data uint64
}

func (s *optUint64) Number() OptionNumber { return s.Type }
func (s *optUint64) Uint64Data() uint64   { return s.Data }

func (s *optUint64) Bytes() []byte {
	var bb bytes.Buffer
	_ = binary.Write(&bb, binary.BigEndian, s.Data)

	var ret = bb.Bytes()
	for i := 0; i < len(ret); i++ {
		if ret[i] != 0 {
			return ret[i:]
		}
	}
	return nil
}

func (s *optUint64) String() string {
	return fmt.Sprintf("[%v: %d/0x%08X]", s.Type, s.Data, s.Data)
}

//
//
//

func NewOptBlockNForDecoding(N int, data []byte) Option {
	var num uint64
	var more bool
	if len(data) > 1 {
		for i := 0; i < len(data)-1; i++ {
			num = num*0x100 + uint64(data[i])
		}
		num <<= 4
	}
	b := data[len(data)-1]
	num += uint64(b >> 4)
	if ((b >> 3) & 1) == 1 {
		more = true
	}
	szx := b & 0x7
	s := &OptBlockN{
		N:    N,
		Num:  num,
		More: more,
		Szx:  szx,
	}
	return s
}

type OptBlockN struct {
	optBase
	N    int
	Num  uint64
	More bool
	Szx  byte
}

func (s *OptBlockN) Number() OptionNumber {
	if s.N == 1 {
		return OptionNumberBlock1
	}
	return OptionNumberBlock2
}

func (s *OptBlockN) SizeInBytes() int { return 1 << (s.Szx + 4) }

func (s *OptBlockN) Bytes() []byte {
	var bb bytes.Buffer
	hiBytes, loByte := bigEndianHalfEncodeUint64(s.Num)
	M := byte(0)
	if s.More {
		M = 1 << 3
	}
	vv := (loByte << 4) | M | s.Szx
	_, _ = bb.Write(hiBytes)
	_ = bb.WriteByte(vv)
	return bb.Bytes()
}

func (s *OptBlockN) String() string {
	return fmt.Sprintf("[Block%d #%d: szx=%v,more=%v]", s.N, s.Num, 1<<(4+s.Szx), s.More)
}

//
//
//

func NewOptETag(optNum OptionNumber, eTag uint64) Option {
	return &optETag{Type: optNum, ETag: eTag}
}

type optETag struct {
	optBase
	Type OptionNumber
	ETag uint64
}

func (s *optETag) Uint64Data() uint64   { return s.ETag }
func (s *optETag) Number() OptionNumber { return s.Type }

func (s *optETag) Bytes() []byte {
	var bb bytes.Buffer
	_ = binary.Write(&bb, binary.BigEndian, s.ETag)

	var ret = bb.Bytes()
	for i := 0; i < len(ret); i++ {
		if ret[i] != 0 {
			return ret[i:]
		}
	}
	return nil
}

func (s *optETag) String() string {
	return fmt.Sprintf("[%v: %08X]", s.Type, s.ETag)
}

//
//
//

func NewOptMediaType(optNum OptionNumber, mt MediaType) Option {
	return &optMediaType{Type: optNum, MediaType: mt}
}

type optMediaType struct {
	optBase
	Type OptionNumber
	MediaType
}

func (s *optMediaType) MediaTypeData() MediaType { return s.MediaType }
func (s *optMediaType) Number() OptionNumber     { return s.Type }

func (s *optMediaType) Bytes() []byte {
	var bb bytes.Buffer
	_ = binary.Write(&bb, binary.BigEndian, s.MediaType)

	var ret = bb.Bytes()
	for i := 0; i < len(ret); i++ {
		if ret[i] != 0 {
			return ret[i:]
		}
	}
	return nil
}

func (s *optMediaType) String() string {
	return fmt.Sprintf("[%v: %q]", s.Type, s.MediaType.String())
}

//
//
//

func decodeUint16(data []byte) (r uint16) {
	for _, d := range data {
		r *= 0x100
		r += uint16(d)
	}
	return
}

func decodeUint32(data []byte) (r uint32) {
	for _, d := range data {
		r *= 0x100
		r += uint32(d)
	}
	return
}

func decodeUint64(data []byte) (r uint64) {
	for _, d := range data {
		r *= 0x100
		r += uint64(d)
	}
	return
}
