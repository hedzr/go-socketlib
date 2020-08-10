package message

import (
	"fmt"
	"io"
)

type Opt interface {
	Number() OptionNumber
	Bytes() []byte
	fmt.Stringer

	SetBytes(data []byte)
	WriteOptDeltaAndLength(w io.Writer, oneByteBuffer []byte, delta OptionNumber, length int) (err error)
}

type optionDecoder struct {
	decoders map[OptionNumber]func(optNum OptionNumber, data []byte) Opt
	err      error
}

func (s *optionDecoder) Error() error {
	return s.err
}

func (s *optionDecoder) Decode(optNum OptionNumber, data []byte) (opt Opt) {
	if s.decoders != nil {
		if d, ok := s.decoders[optNum]; ok {
			opt = d(optNum, data)
			return
		}
	}

	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionIfMatch(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionUriHost(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionETag(optNum OptionNumber, data []byte) (opt Opt) {
	var eTag = decodeUint64(data)
	opt = NewOptETag(eTag)
	return
}

func (s *optionDecoder) decodeOptionIfNoneMatch(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionObserve(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionUriPort(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionLocationPath(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionUriPath(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionContentFormat(optNum OptionNumber, data []byte) (opt Opt) {
	n := decodeUint16(data)
	opt = NewOptContentFormat(MediaType(n))
	return
}

func (s *optionDecoder) decodeOptionMaxAge(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionUriQuery(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionAccept(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionLocationQuery(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionBlock2(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewOptBlockNForDecoding(2, data)
	return
}

func (s *optionDecoder) decodeOptionBlock1(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewOptBlockNForDecoding(1, data)
	return
}

func (s *optionDecoder) decodeOptionSize2(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionProxyUri(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionProxyScheme(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}

func (s *optionDecoder) decodeOptionSize1(optNum OptionNumber, data []byte) (opt Opt) {
	opt = NewAnyOpt(optNum, data)
	return
}
