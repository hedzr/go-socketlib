package pi

import (
	"bytes"
	"encoding/binary"
	"sync"
)

const maxPackageSize = 65536
const initialBufferSize = 4096

func NewLeadBytes(leadingMagics []byte) *leadBytesS {
	s := &leadBytesS{
		leadingMagics: leadingMagics,
		// decodeCache:   poolLeadByteDecode.Get().([]byte),
		// encodeCache:   poolLeadByteEncode.Get().([]byte),
	}
	return s
}

var poolLeadByteDecode = sync.Pool{New: func() any { return make([]byte, 0, initialBufferSize) }}
var poolLeadByteEncode = sync.Pool{New: func() any { return make([]byte, 0, initialBufferSize) }}

type leadBytesS struct {
	leadingMagics []byte
	decodeCache   []byte
	encodeCache   []byte
}

func (s *leadBytesS) OnEncode(body []byte) (data []byte, err error) {
	if ld := len(body); ld > 0 {
		defer poolLeadByteDecode.Put(s.decodeCache)
		l := len(s.leadingMagics)
		s.encodeCache = poolLeadByteEncode.Get().([]byte)[:l]
		copy(s.encodeCache, s.leadingMagics)
		s.encodeCache = binary.AppendVarint(s.encodeCache, int64(ld))
		s.encodeCache = append(s.encodeCache, body...)
		data = s.encodeCache
	}
	return
}

func (s *leadBytesS) OnDecode(data []byte, ch chan<- []byte) (processed bool, err error) {
	if ld := len(data); ld > 0 {
		defer poolLeadByteDecode.Put(s.decodeCache)
		s.decodeCache = append(poolLeadByteDecode.Get().([]byte)[:0], data...)
		cacheLen := len(s.decodeCache)
		if l := len(s.leadingMagics); cacheLen > l {
		nextPackage:
			if bytes.Compare(s.leadingMagics, s.decodeCache[:l]) == 0 {
				var length int64
				var ate int
				length, ate = binary.Varint(s.decodeCache[l:])
				if ate > 0 && length < maxPackageSize {
					processed = true
					begin := l + ate
					end := int(length) + begin
					if end <= cacheLen {
						if ch != nil {
							ch <- s.decodeCache[begin:end]
						}
						s.decodeCache = s.decodeCache[end:]
						goto nextPackage
					}
				}
			}
		}
	}
	return
}
