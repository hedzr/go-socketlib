package message

import "bytes"

func NewBlock2(blockNum int, hasMore bool, sizeInBytes uint, contents []byte) Block {
	return NewBlock1(blockNum, hasMore, sizeInBytes, contents)
}

func NewBlock1(blockNum int, hasMore bool, sizeInBytes uint, contents []byte) Block {
	s := &anyBlock{
		Num: blockNum,
	}

	var err error
	var w bytes.Buffer

	if hasMore {
		s.hasMore = 1 << 3
	}
	s.szx = RoundedLog2(sizeInBytes) - 4

	if blockNum > 16 {
		print(blockNum)
	}

	hiBytes, loByte := bigEndianHalfEncodeUint64(uint64(blockNum))
	vv := (loByte << 4) | s.hasMore | s.szx
	// s.err = binary.Write(&w, binary.BigEndian, vv)
	_, err = w.Write(hiBytes)
	err = w.WriteByte(vv)

	if err == nil {
		if contents != nil {
			_, err = w.Write(contents)
		}
		if err == nil {
			s.data = w.Bytes()
		}
	}

	return s
}

type Block interface {
	Number() int
	Size() int
	Bytes() []byte
}

type anyBlock struct {
	Num     int
	szx     byte
	hasMore byte
	data    []byte
}

func (s *anyBlock) Number() int {
	return s.Num
}

func (s *anyBlock) Size() int {
	return len(s.data)
}

func (s *anyBlock) Bytes() []byte {
	return s.data
}

func bigEndianHalfEncodeUint64(num uint64) (hiBytes []byte, loByte byte) {
	loByte = byte(num & 0x0f)
	num >>= 4
	for i := 7; i >= 0; i-- {
		z := byte(num >> (i * 8))
		if z == 0 {
			if len(hiBytes) == 0 {
				continue
			}
		}
		hiBytes = append(hiBytes, z)
	}
	return
}

// fast log2(v)
func RoundedLog2(v uint) byte {
	var r, shift uint

	if v > 0xFFFF {
		r = 16
		v >>= r
	}

	if v > 0xFF {
		shift = 8
		v >>= shift
		r |= shift
	}

	if v > 0xF {
		shift = 4
		v >>= shift
		r |= shift
	}

	if v > 0x3 {
		shift = 2
		v >>= shift
		r |= shift
	}

	t := v >> 1
	if t > 0 {
		r |= t
	}

	return byte(r)
}
