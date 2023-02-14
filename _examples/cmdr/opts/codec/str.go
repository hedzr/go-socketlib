package codec

import (
	"bytes"
	"gopkg.in/hedzr/errors.v3"
)

func EncodeString(str string) []byte {
	var buf bytes.Buffer
	data := []byte(str)
	ll := len(data)
	if ll < 32 {
		_l := EncodeVarInt64(-int64(ll))
		buf.Write(_l)
		buf.Write(data)
	} else {
		if b, err := compressByFlate(data); err == nil {
			length := len(b) // + len(l)
			dl := EncodeVarInt64(int64(length))
			buf.Write(dl)
			buf.Write(b)
		}
	}
	return buf.Bytes()
}

func DecodeString(data []byte) (str string, nRead int, err error) {
	var (
		x  int64
		n  int
		ok bool
		r  []byte
	)
	x, n, _, ok = DecodeVarInt(data)
	if ok {
		if x < 0 {
			pos := nRead + n
			nRead = pos + int(-x)
			str = string(data[pos:nRead])
		} else {
			nRead += n
			//y, n, _, ok = DecodeVarIntU(data[nRead:])
			//if ok {
			//	nRead += n
			r, n, err = decompressByDeflate(data[nRead:])
			if err == nil {
				str = string(r)
				if n != len(str) {
					err = errors.New("unexpect string length found (want %v but got %v)", n, len(str))
				} else {
					nRead += int(x)
				}
			}
		}
	}
	return
}
