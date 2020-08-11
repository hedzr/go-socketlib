package message

import (
	"bytes"
	"fmt"
)

type Payload interface {
	ContentFormat() MediaType
	Bytes() []byte
	Size() int
	Set(data []byte)
}

func NewPayload(data []byte) Payload {
	return NewPayloadWith(data, TextPlain)
}

func NewPayloadWith(data []byte, contentFormat MediaType) Payload {
	s := &anyPayload{MediaType: contentFormat}
	copy(s.Data, data)
	return s
}

func payloadCreate(data []byte, startPosition int, mt MediaType, options []Option) (p Payload) {
	if startPosition < len(data) {
		switch mt {
		case AppLinkFormat: // rfc6690
			s := &payloadLinkFormat{}
			s.originalBin.Write(data[startPosition+1:])

			opt := FindOption(OptionNumberBlock2, options)
			if optBlock2, ok := opt.(*OptBlockN); ok {
				if optBlock2.More == false {
					var err error
					s.res, err = s.Parse(s.originalBin.String())
					if err != nil {
						logger.Errorf("link format parse failed: %v", err)
					}
				}
			}
			p = s

		default:
			s := &anyPayload{MediaType: mt}
			s.Data = make([]byte, len(data)-startPosition-1)
			copy(s.Data, data[startPosition+1:])
			p = s
		}
	}
	return
}

type payloadLinkFormat struct {
	LinkFormatParser
	res         []*lfResource
	originalBin bytes.Buffer
}

func (s *payloadLinkFormat) ContentFormat() MediaType {
	return AppLinkFormat
}

func (s *payloadLinkFormat) Bytes() []byte {
	if s.originalBin.Len() > 0 {
		return s.originalBin.Bytes()
	}

	return s.LinkFormatParser.ToBytes(s.res)
}

func (s *payloadLinkFormat) Size() int {
	if s == nil {
		return 0
	}
	t := s.Bytes()
	return len(t)
}

func (s *payloadLinkFormat) Set(data []byte) {
	panic("implement me")
}

func (s *payloadLinkFormat) String() string {
	return string(s.Bytes())
}

//
//
//

type anyPayload struct {
	MediaType
	Data []byte
}

func (s *anyPayload) String() string {
	if s.Data != nil {
		switch s.MediaType {
		case TextPlain, AppXML, AppJSON, AppJSONPatch,
			AppJSONMergePatch, AppCoapGroup, AppLwm2mJSON,
			AppLinkFormat:
			return string(s.Data)
		}
		return fmt.Sprintf("% x", s.Data)
	}
	return ""
}

func (s *anyPayload) ContentFormat() MediaType {
	return s.MediaType
}

func (s *anyPayload) Bytes() []byte {
	return s.Data
}

func (s *anyPayload) Size() int {
	if s == nil {
		return 0
	}
	return len(s.Data)
}

func (s *anyPayload) Set(data []byte) {
	lSelf, ll := len(s.Data), len(data)
	if lSelf < ll {
		s.Data = make([]byte, ll)
	} else if lSelf > ll {
		s.Data = s.Data[:ll]
	}
	copy(s.Data, data)
}
