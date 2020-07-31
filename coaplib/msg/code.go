package msg

import "fmt"

type Code uint8
type CodeCategory uint8
type CodeDetails uint8

func (c Code) Category() CodeCategory {
	return CodeCategory(c >> 5)
}

func (c Code) Details() CodeDetails {
	return CodeDetails(c & 0x1f)
}

func (c Code) String() string {
	a, b := c>>5, c&0x1f
	return fmt.Sprintf("%d.%02d", a, b)
}
