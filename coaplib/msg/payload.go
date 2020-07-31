package msg

import "fmt"

type Payload interface {
}

func payloadCreate(data []byte, i int) (p Payload) {
	if i < len(data) {
		s := &anyPayload{}
		copy(s.Data, data[i:])
		return s
	}
	return
}

type anyPayload struct {
	Data []byte
}

func (s *anyPayload) String() string {
	if s.Data != nil {
		return fmt.Sprintf("% x", s.Data)
	}
	return ""
}
