package pi

import (
	"github.com/hedzr/go-socketlib/coaplib/message"
	"net/url"
	"strings"
)

type Builder struct {
	msg                  *message.Message
	uri                  *url.URL
	proxyURI             string
	proxyScheme          string
	eTag                 uint64
	ifMatch, ifNoneMatch uint64
	maxAge               uint32
	locationPath         string
	locationQuery        string
	accept               message.MediaType
	size1, size2         uint64
	blocks1              []message.Block
	blocks2              []message.Block
	lastBlockNum         int
	err                  error
	optionProcMap        map[message.OptionNumber]func() []message.Option
	optNum               message.OptionNumber
	//lastOptNum           message.OptionNumber
}

var seq uint32 = 101

func (s *Builder) Build() (out *message.Message) {
	out = s.msg
	//s.msg = new(message.Message)

	// var optNum message.OptionNumber
	for _, k := range message.OptionNumberSortedKeys {
		s.optNum = k

		proc := s.optionProcMap[s.optNum]
		if proc == nil {
			continue
		}

		opt := proc()
		if opt != nil {
			// s.lastOptNum = k
			for _, oo := range opt {
				if oo != nil {
					s.msg.Options = append(s.msg.Options, oo)
				}
			}
		}
	}

	return
}

func (s *Builder) addOptionUint64(optNum message.OptionNumber, i uint64) (opt message.Option) {
	if i >= 0 {
		opt = message.NewUint64Opt(optNum, i)

		//length := s.msg.CalcDigitLen(i)
		//delta := s.optNum - s.lastOptNum
		//
		//var bb bytes.Buffer
		//
		//s.err = opt.WriteOptDeltaAndLength(&bb, delta, length)
		//if s.err != nil {
		//	return
		//}
		//
		//var bt bytes.Buffer
		//s.err = binary.Write(&bt, binary.BigEndian, i)
		//if s.err != nil {
		//	return
		//}
		//
		//bb.Write(bt.Bytes()[:length])
		//
		//opt.SetBytes(bb.Bytes())
		//s.lastOptNum = s.optNum
	}
	return
}

func (s *Builder) addOptionString(str string) (opt message.Option) {
	if len(str) > 0 {
		opt = message.NewStringOpt(s.optNum, str)
	}
	return

	//length := len(str)
	//delta := s.optNum - s.lastOptNum
	//
	//if length < message.CoapOptionDefs[s.optNum].MinLen || length > message.CoapOptionDefs[s.optNum].MaxLen {
	//	s.err = errors.New("not in length range: %v..%v", message.CoapOptionDefs[s.optNum].MinLen, message.CoapOptionDefs[s.optNum].MaxLen)
	//	return
	//}
	//
	//if length <= 0 {
	//	return
	//}
	//
	//var bb bytes.Buffer
	//
	//opt = message.NewAnyOpt(s.optNum, nil)
	//
	//s.err = opt.WriteOptDeltaAndLength(&bb, delta, length)
	//if s.err != nil {
	//	return
	//}
	//
	//_, s.err = bb.WriteString(str)
	//if s.err != nil {
	//	return
	//}
	//
	//opt.SetBytes(bb.Bytes())
	//s.lastOptNum = s.optNum
	//
	//return
}

func (s *Builder) addOptionIfMatch() (opt []message.Option) {
	if s.ifMatch > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, s.ifMatch))
	}
	return
}

func (s *Builder) addOptionUriHost() (opt []message.Option) {
	if s.uri == nil {
		return
	}
	opt = append(opt, s.addOptionString(s.uri.Hostname()))
	return
}

func (s *Builder) addOptionETag() (opt []message.Option) {
	if s.eTag > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, s.eTag))
	}
	return
}

func (s *Builder) addOptionIfNoneMatch() (opt []message.Option) {
	if s.ifNoneMatch > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, s.ifNoneMatch))
	}
	return
}

func (s *Builder) addOptionUriPort() (opt []message.Option) {
	if s.uri == nil {
		return
	}
	if len(s.uri.Port()) > 0 {
		opt = append(opt, s.addOptionString(s.uri.Port()))
	}
	return
}

func (s *Builder) addOptionLocationPath() (opt []message.Option) {
	if len(s.locationPath) > 0 {
		opt = append(opt, s.addOptionString(s.locationPath))
	}
	return
}

func (s *Builder) addOptionUriPath() (opt []message.Option) {
	if s.uri == nil {
		return
	}

	for _, p := range strings.Split(s.uri.Path, "/") {
		opt = append(opt, s.addOptionString(p))
	}
	return
}

func (s *Builder) addOptionContentFormat() (opt []message.Option) {
	if s.msg.Payload != nil {
		opt = append(opt, s.addOptionUint64(s.optNum, uint64(s.msg.Payload.ContentFormat())))
	}
	return
}

func (s *Builder) addOptionMaxAge() (opt []message.Option) {
	if s.maxAge != 60 && s.maxAge > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, uint64(s.maxAge)))
	}
	return
}

func (s *Builder) addOptionUriQuery() (opt []message.Option) {
	if s.uri == nil {
		return
	}

	for k, vs := range s.uri.Query() {
		for _, v := range vs {
			opt = append(opt, s.addOptionString(url.QueryEscape(k)+"="+url.QueryEscape(v)))
		}
	}
	//opt = append(opt, s.addOptionString(s.uri.RawQuery))
	return
}

func (s *Builder) addOptionAccept() (opt []message.Option) {
	if s.accept != message.MediaTypeUndefined {
		opt = append(opt, s.addOptionUint64(s.optNum, uint64(s.accept)))
	}
	return
}

func (s *Builder) addOptionLocationQuery() (opt []message.Option) {
	if len(s.locationQuery) > 0 {
		values, err := url.ParseQuery(s.locationQuery)
		if err == nil {
			for k, vs := range values {
				for _, v := range vs {
					opt = append(opt, s.addOptionString(url.QueryEscape(k)+"="+url.QueryEscape(v)))
				}
			}
		}
		//opt = append(opt, s.addOptionString(s.locationQuery))
	}
	return
}

func (s *Builder) addOptionProxyUri() (opt []message.Option) {
	if len(s.proxyURI) > 0 {
		opt = append(opt, s.addOptionString(s.proxyURI))
	}
	return
}

func (s *Builder) addOptionProxyScheme() (opt []message.Option) {
	if len(s.proxyScheme) > 0 {
		opt = append(opt, s.addOptionString(s.proxyScheme))
	}
	return
}

func (s *Builder) addOptionSize1() (opt []message.Option) {
	if s.size1 > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, s.size1))
	}
	return
}

func (s *Builder) addOptionSize2() (opt []message.Option) {
	if s.size2 > 0 {
		opt = append(opt, s.addOptionUint64(s.optNum, s.size2))
	}
	return
}

func (s *Builder) addOptionBlock1() (opt []message.Option) {
	if s.lastBlockNum < len(s.blocks1) {
		blockNum := s.lastBlockNum
		s.lastBlockNum++
		block := s.blocks1[blockNum]

		//length := block.Size()
		//delta := s.optNum - s.lastOptNum
		//
		//var bb bytes.Buffer
		//
		//s.writeOptDeltaAndLength(&bb, delta, length)
		//if s.err != nil {
		//	return
		//}
		//
		//bb.Write(block.Bytes())

		opt = append(opt, message.NewAnyOpt(s.optNum, block.Bytes()))
	}
	return
}

//func (s *Builder) setBlock1ToPayload() (opt message.Option) {
//	if s.lastBlockNum < len(s.blocks) {
//		blockNum, hasMore := s.lastBlockNum, byte(0)
//
//		if s.lastBlockNum < len(s.blocks) {
//			hasMore = byte(1) << 3
//		}
//
//		var w bytes.Buffer
//
//		szx := RoundedLog2(block.Size()) - 4
//		hiBytes, loByte := bigEndianHalfEncodeUint64(uint64(blockNum))
//		vv := (loByte << 4) | hasMore | szx
//		// s.err = binary.Write(&w, binary.BigEndian, vv)
//		_, s.err = w.Write(hiBytes)
//		s.err = w.WriteByte(vv)
//
//		if s.err == nil {
//			// s.msg.Payload.Set(block.Content())
//			opt = message.NewBlock(w.Bytes(), blockNum)
//		}
//	}
//	return
//}

func (s *Builder) setBlock2ToOptions() (opt []message.Option) {
	//opt = s.addOptionUint64(uint64(s.size2))
	for _, block := range s.blocks2 {
		//length := block.Size()
		//delta := s.optNum - s.lastOptNum
		//
		//var bb bytes.Buffer
		//
		//s.writeOptDeltaAndLength(&bb, delta, length)
		//if s.err != nil {
		//	return
		//}
		//
		//bb.Write(block.Bytes())

		opt = append(opt, message.NewAnyOpt(s.optNum, block.Bytes()))
	}
	return
}

func (s *Builder) blocksToOptions(blocks []message.Block, optNum message.OptionNumber) (opt []message.Option) {
	//opt = s.addOptionUint64(uint64(s.size2))
	for _, block := range blocks {
		o := s.blockToOption(block, optNum)
		opt = append(opt, o)
	}
	return
}

func (s *Builder) blockToOption(block message.Block, optNum message.OptionNumber) (opt message.Option) {
	//length := block.Size()
	//delta := s.optNum - s.lastOptNum
	//
	//var bb bytes.Buffer
	//
	//s.writeOptDeltaAndLength(&bb, delta, length)
	//if s.err != nil {
	//	return
	//}
	//
	//bb.Write(block.Bytes())

	opt = message.NewAnyOpt(optNum, block.Bytes())
	return
}
