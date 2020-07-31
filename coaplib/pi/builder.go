package pi

import "github.com/hedzr/go-socketlib/coaplib/msg"

type Builder struct {
}

func NewBuilder() *Builder {
	return &Builder{}
}

func (s *Builder) Build() (out *msg.Message) {
	return
}

func (s *Builder) With() *Builder {
	return s
}

func (s *Builder) With2() *Builder {
	return s
}

func (s *Builder) With3() *Builder {
	return s
}
