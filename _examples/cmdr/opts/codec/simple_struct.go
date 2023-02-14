package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// EncodeStruct serializes a simple flat struct as a byte slice.
//
//     type ts struct {
//         Pi  float64
//         I64 int64
//     }
//     c := ts{3.141592653589794, -1}
//     buf := EncodeStruct(c)
//     buf := EncodeStruct(&c)
func EncodeStruct(s interface{}) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, s)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

// DecodeStruct deserializes the byte slices into a simple flat struct.
//
//     var tgt *ts = new(ts)
//     nRead, err := DecodeStruct(buf, tgt)
//
// The simple flat struct: without pointer, string fields.
// *anotherStruct, map, string are invalid field types.
func DecodeStruct(data []byte, ret interface{}) (n int, err error) {
	r := bytes.NewReader(data)
	if err = binary.Read(r, ByteOrder, ret); err == nil {
		var pos int64
		pos, err = r.Seek(0, io.SeekCurrent)
		n = int(pos)
	}
	return
}
