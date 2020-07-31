package msg

import "fmt"

type Code uint8
type CodeCategory uint8
type CodeDetails uint8

func (i Code) Category() CodeCategory {
	return CodeCategory(i >> 5)
}

func (i Code) Details() CodeDetails {
	return CodeDetails(i & 0x1f)
}

func (i Code) StringEx() string {
	a, b := i>>5, i&0x1f
	return fmt.Sprintf("%d.%02d", a, b)
}
