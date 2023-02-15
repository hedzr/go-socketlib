package pibufio

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hedzr/log"
	"gopkg.in/hedzr/errors.v3"

	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/codec"
)

const (
	maxCachedPackages = 4096
	maxBufferSize     = 8 * 1024 * 1024 // 8MB
)

func New() Queue {
	var traceEnabled = log.GetDebugMode() || log.GetTraceMode()
	return &iobuf{
		err:              nil,
		packageExtractor: &pe{},
		chReceiver:       make(chan []byte, maxCachedPackages),
		traceEnabled:     traceEnabled,
	}
}

type Queue interface {
	Enqueue(data []byte) (err error)
	PkgReceived() chan []byte
	TryExtractPackages()
	Error() error
}

type iobuf struct {
	bytes.Buffer
	rw                sync.RWMutex
	err               error
	packageExtractor  PackageExtractor
	chReceiver        chan []byte
	closed            int32
	traceEnabled      bool
	cachedLengthField int
}

func (s *iobuf) Close() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		close(s.chReceiver)
	}
}

func (s *iobuf) SetTraceEnabled(b bool)   { s.traceEnabled = b }
func (s *iobuf) Error() error             { return s.err }
func (s *iobuf) PkgReceived() chan []byte { return s.chReceiver }

func (s *iobuf) TryExtractPackages() {
	for s.try() {
		time.Sleep(1 * time.Nanosecond)
	}

	defer s.rw.Unlock()
	s.rw.Lock()
	if s.Len() == 0 && s.Cap() > maxBufferSize {
		log.Debugf("pibufio: buffer reset since its capacity is over %v", s.Cap())
		s.Reset()
	}
}

func (s *iobuf) Enqueue(data []byte) error {
	if atomic.LoadInt32(&s.closed) == 0 {
		if s.err == nil {
			s.enqueue(data)
			if s.err == nil {
				s.try()
			}
		} else if s.traceEnabled {
			log.Errorf("iobuf.Enqueue failed: %v", s.err)
		}
	}
	return s.err
}

func (s *iobuf) enqueue(data []byte) {
	defer s.rw.Unlock()
	s.rw.Lock()
	_, s.err = s.Buffer.Write(data)
	if s.traceEnabled && s.err != nil {
		log.Errorf("iobuf.enqueue failed: %v", s.err)
	}
}

func (s *iobuf) try() (extracted bool) {
	defer s.rw.RUnlock()
	s.rw.RLock()
	for {
		if s.cachedLengthField > 0 {
			if s.Buffer.Len() >= s.cachedLengthField {
				extracted = s.got(s.cachedLengthField)
				s.cachedLengthField = 0
				return
			}
		}

		if c, err := s.Buffer.ReadByte(); err != nil || c != 0xaa {
			s.err = err
			break
		}

		if c, err := s.Buffer.ReadByte(); err != nil || c != 0x55 {
			// if gapWord (2 bytes) not found, they will be
			// discarded. And we'll step on looking for the
			// next location of the gapWord.
			if s.traceEnabled {
				log.Warnf("iobuf.try's discarding single 0xaa, ReadUvarint() failed: %v", s.err)
			}
			continue
		}

		var length uint16
		err := binary.Read(&s.Buffer, codec.ByteOrder, &length)
		if err == nil {
			if s.Buffer.Len() >= int(length) {
				extracted = s.got(int(length))
				s.err = err
				break
			}

			s.cachedLengthField = int(length)
			break
		}

		if s.traceEnabled {
			// if gapWord (2 bytes) not found, they will be
			// discarded. And we'll step on looking for the
			// next location of the gapWord.
			log.Warnf("iobuf.try's discarding single 0xaa, ReadUvarint() failed: %v", s.err)
		}

		// length, s.err = binary.ReadUvarint(&s.Buffer)
		// if s.err == nil {
		//	if s.Buffer.Len() > int(length) {
		//		data := s.Next(int(length))
		//		s.chReceiver <- data
		//		extracted = true
		//		return
		//	}
		//
		//	count := binary.PutUvarint(internalTmpBuf, length)
		//	for i := 0; i < count; i++ {
		//		s.err = s.Buffer.UnreadByte()
		//	}
		//	s.err = s.Buffer.UnreadByte()
		//	s.err = s.Buffer.UnreadByte()
		//	break
		//
		// } else if s.traceEnabled {
		//	// if gapWord (2 bytes) not found, they will be
		//	// discarded. And we'll step on looking for the
		//	// next location of the gapWord.
		//	log.Warnf("iobuf.try's discarding single 0xaa, ReadUvarint() failed: %v", s.err)
		// }
	}

	if errors.Is(s.err, io.EOF) {
		s.err = nil
	}
	return
}

func (s *iobuf) got(howMany int) (extracted bool) {
	data := s.Next(howMany)
	log.Debugf("pkg received [len=%v,cap=%v]: %v", s.Len(), s.Cap(), data)
	s.chReceiver <- data
	extracted = true
	return
}

// var internalTmpBuf = make([]byte, 16)

type PackageExtractor interface {
	Try(buf *bytes.Buffer)
}

type pe struct {
}

func (s *pe) Try(buf *bytes.Buffer) {
	buf.Bytes()
}
