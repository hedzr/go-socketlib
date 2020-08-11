package pi

import (
	"github.com/hedzr/cmdr"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"gopkg.in/hedzr/errors.v2"
	"net/url"
	"sync/atomic"
)

// NewBuilder make a builder instance.
//
// Usage:
//
//     builder := NewBuilder()
//     msg := builder.
//       WithType(message.CON).
//       WithCode(message.MethodCodePOST).
//       WithURIString("coap://coap.me/large").
//       Build()
//
// Reuse a Builder:
//
//     builder := NewBuilder()
//     //...
//     msg := builder.Build()
//     //...
//     builder.From(msg, true)
//     msg := builder.Build()
//     //...
//     builder.NewBase("coap://coap.me").WithURIPath("/test")
//     msg := builder.Build()
//
// Duplicate a new builder from exists:
//
//     builder := NewBuilder()
//     builderNew := builder.Clone()
//
// See also Builder.Reset(), Builder.Clone(), Builder.From(),
// Builder.NewBase(), ....
//
func NewBuilder() *Builder {
	return newBuilder()
}

func createOptionProcMap(b *Builder) (m map[message.OptionNumber]func() []message.Option) {
	m = map[message.OptionNumber]func() []message.Option{
		0:                                 nil,
		message.OptionNumberIfMatch:       b.addOptionIfMatch,
		message.OptionNumberURIHost:       b.addOptionUriHost,
		message.OptionNumberETag:          b.addOptionETag,
		message.OptionNumberIfNoneMatch:   b.addOptionIfNoneMatch,
		message.OptionNumberObserve:       nil,
		message.OptionNumberURIPort:       b.addOptionUriPort,
		message.OptionNumberLocationPath:  b.addOptionLocationPath,
		message.OptionNumberURIPath:       b.addOptionUriPath,
		message.OptionNumberContentFormat: b.addOptionContentFormat,
		message.OptionNumberMaxAge:        b.addOptionMaxAge,
		message.OptionNumberURIQuery:      b.addOptionUriQuery,
		message.OptionNumberAccept:        b.addOptionAccept,
		message.OptionNumberLocationQuery: b.addOptionLocationQuery,
		message.OptionNumberBlock2:        b.setBlock2ToOptions,
		message.OptionNumberBlock1:        b.addOptionBlock1, //setBlock1ToPayload,
		message.OptionNumberSize2:         b.addOptionSize2,
		message.OptionNumberProxyURI:      b.addOptionProxyUri,
		message.OptionNumberProxyScheme:   b.addOptionProxyScheme,
		message.OptionNumberSize1:         b.addOptionSize1,
		message.OptionNumberR128:          nil,
		message.OptionNumberR132:          nil,
		message.OptionNumberR136:          nil,
		message.OptionNumberR140:          nil,
		message.OptionNumberNoResponse:    nil,
	}
	return
}

func newBuilder() *Builder {
	b := &Builder{
		msg: message.New(message.MethodCodeGET, message.WithMessageID(seq)),
		// token: rand.Uint64(),
		// seq: 101,
		accept: message.MediaTypeUndefined,
	}
	b.optionProcMap = createOptionProcMap(b)
	return b
}

func (s *Builder) Clone() *Builder {
	b := new(Builder)
	cmdr.Clone(s, b)
	return b
}

func (s *Builder) Error() error {
	return s.err
}

func (s *Builder) From(msg *message.Message, removeBlock2Opt bool) *Builder {
	ret := s.Reset()
	ret.msg = msg
	// ret.msg.MessageID = uint16(atomic.AddUint32(&seq, -1))
	ret.msg.MessageID = uint16(atomic.LoadUint32(&seq))
	//for _, o := range ret.msg.Options {
	//	ret.lastOptNum = o.Number()
	//}
	return ret
}

func (s *Builder) NewBase(baseUrl string) *Builder {
	ret := s.Reset()
	return ret.WithURI2(url.Parse(baseUrl))
}

//
func (s *Builder) Reset() *Builder {
	var token uint64
	if s.msg != nil {
		token = s.msg.Token
	}

	s.msg = message.New(message.MethodCodeGET,
		message.WithMessageID(atomic.AddUint32(&seq, 1)),
		message.WithToken(token),
	)

	s.uri = nil
	s.proxyURI = ""
	s.proxyScheme = ""
	s.eTag = 0
	s.ifMatch = 0
	s.ifNoneMatch = 0
	s.maxAge = 0
	s.accept = message.MediaTypeUndefined
	s.locationPath = ""
	s.locationQuery = ""
	s.size1 = 0
	s.size2 = 0
	s.blocks1 = nil
	s.blocks2 = nil
	s.lastBlockNum = 0
	//s.lastOptNum = 0
	s.optNum = 0
	s.err = nil

	return s
}

func (s *Builder) WithOnACK(fn message.OnACKHandler) *Builder {
	if fn != nil {
		s.msg.WithOnACK(fn)
	}
	return s
}

func (s *Builder) WithOnEvent(fn message.OnEventHandler) *Builder {
	if fn != nil {
		s.msg.WithOnEvent(fn)
	}
	return s
}

func (s *Builder) WithRegister(reg int) *Builder {
	if reg == 0 || reg == 1 {
		o1 := s.addOptionUint64(message.OptionNumberObserve, uint64(reg))
		o2 := s.addOptionUint64(message.OptionNumberAccept, uint64(message.TextPlain))
		s.msg.Options = append(s.msg.Options, o1, o2)
	}
	return s
}

func (s *Builder) WithMessageOptions(options ...message.Option) *Builder {
	for _, o := range options {
		if o != nil {
			s.msg.Options = append(s.msg.Options, o)
		}
	}
	return s
}

// WithType:
//     Type: message.CON, message.NON, message.ACK, message.RST
func (s *Builder) WithType(typ message.Type) *Builder {
	s.msg.Type = typ
	return s
}

// WithCode:
//     MethodCode    : message.MethodCodeGET, message.MethodCodePOST, message.MethodCodePUT, message.MethodCodeDELETE, ...
//     ResponseCode  : message.ResponseCodeCreated, ...
//
func (s *Builder) WithCode(code message.Code) *Builder {
	s.msg.Code = code
	return s
}

func (s *Builder) WithMessageID(mid uint16) *Builder {
	s.msg.MessageID = mid
	return s
}

func (s *Builder) WithToken(token uint64) *Builder {
	s.msg.SetToken(token)
	// s.token = token
	return s
}

// WithURIPath needs WithURI() / WithURI2() / NewBase() first.
//
// Sample:
//     builder.WithURIString("coap://coap.me/large")...Build()
//     ...
//     builder.
//         Reset().
//         WithURIString("coap://coap.me/temp")
//     builder.
//         WithURIPath("/test")
//     builder.
//         Build()
//     ...
//     builder.
//         NewBase("coap://coap.me")
//     builder.
//         WithURIPath("/test")
//     builder.
//         Build()
//
func (s *Builder) WithURIPath(path string) *Builder {
	if s.uri == nil {
		s.err = errors.New("Builder.NewBase() needed before WithURIPath()")
		return nil
	}

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	s.uri.Path = path
	return s
}

func (s *Builder) WithURI(uri *url.URL) *Builder {
	s.uri = uri
	return s
}

func (s *Builder) WithURI2(uri *url.URL, err error) *Builder {
	if err == nil {
		s.uri = uri
	}
	return s
}

func (s *Builder) WithURIString(uriString string) *Builder {
	uri, err := url.Parse(uriString)
	if err == nil {
		s.uri = uri
	}
	return s
}

func (s *Builder) WithProxy(proxyScheme, proxyURI string) *Builder {
	switch proxyScheme {
	case "http", "https":
	case "socks", "sock5", "sock4":
	default:
		s.err = errors.New("unknown proxy scheme: %q", proxyScheme)
		return s
	}

	s.proxyScheme = proxyScheme
	s.proxyURI = proxyURI
	return s
}

func (s *Builder) WithIfMatch(i uint64) *Builder {
	s.ifMatch = i
	return s
}

func (s *Builder) WithIfNoneMatch(i uint64) *Builder {
	s.ifNoneMatch = i
	return s
}

func (s *Builder) WithETag(eTag uint64) *Builder {
	s.eTag = eTag
	return s
}

func (s *Builder) WithMaxAge(maxAgeSeconds uint32) *Builder {
	s.maxAge = maxAgeSeconds
	return s
}

func (s *Builder) WithPayload(payload message.Payload) *Builder {
	s.msg.Payload = payload
	return s
}

func (s *Builder) WithAccept(accept message.MediaType) *Builder {
	s.accept = accept
	return s
}

func (s *Builder) WithLocation(locationPath, locationQuery string) *Builder {
	s.locationPath, s.locationQuery = locationPath, locationQuery
	return s
}

func (s *Builder) WithSize1(sz int) *Builder {
	s.size1 = uint64(sz)
	return s
}

func (s *Builder) WithSize2(sz int) *Builder {
	s.size2 = uint64(sz)
	return s
}

func (s *Builder) WithRequestBlock1Num(number int, maxSizeInBytes uint) *Builder {
	block := message.NewBlock2(number, false, maxSizeInBytes, nil)

	for i, o := range s.msg.Options {
		if o.Number() == message.OptionNumberBlock2 {
			s.msg.Options[i] = s.blockToOption(block, o.Number())
			return s
		}
	}

	s.blocks2 = append(s.blocks2, block)
	return s
}

func (s *Builder) WithBlock1(block message.Block) *Builder {
	s.blocks1 = nil
	s.blocks1 = append(s.blocks1, block)
	return s
}

func (s *Builder) WithBlock1Append(block message.Block) *Builder {
	s.blocks1 = append(s.blocks1, block)
	return s
}

func (s *Builder) WithBlock2Append(block message.Block) *Builder {
	s.blocks2 = append(s.blocks2, block)
	return s
}

func (s *Builder) With7() *Builder {
	return s
}
