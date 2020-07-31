package msg

import (
	"fmt"
)

type Opt interface {
}

func newAnyOpt(optType int, data []byte) *anyOpt {
	o := &anyOpt{
		Type: optType,
	}
	o.Data = make([]byte, len(data))
	copy(o.Data, data)
	return o
}

type anyOpt struct {
	Type int
	Data []byte
}

func (s *anyOpt) String() string {
	return fmt.Sprintf("[Type %v: % x]", s.Type, s.Data)
}
