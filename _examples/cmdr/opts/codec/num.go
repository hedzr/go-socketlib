package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/codec/zigzag"
)

func EncodeVarInt64U(x uint64) []byte {
	se := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(se, x)
	return se[:n]
}

func EncodeVarInt32U(x uint32) []byte {
	se := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(se, uint64(x))
	return se[:n]
}

func EncodeVarInt16U(x uint16) []byte {
	se := make([]byte, binary.MaxVarintLen16)
	n := binary.PutUvarint(se, uint64(x))
	return se[:n]
}

func DecodeVarIntU(data []byte) (x uint64, nRead int, smallOrOverflow, ok bool) {
	x, nRead = binary.Uvarint(data)
	if nRead == 0 {
		smallOrOverflow = true
	} else if nRead > 0 {
		ok = true
	} else {
		nRead = -nRead
	}
	return
}

// EncodeVarInt encodes a signed int as sint (in protobuf).
// zigzag the num at first and encode as varint.
func EncodeVarInt(x int64) []byte   { return EncodeVarInt64U(zigzag.Encode(x)) }
func EncodeVarInt64(x int64) []byte { return EncodeVarInt64U(zigzag.Encode(x)) }
func EncodeVarInt32(x int32) []byte { return EncodeVarInt64U(zigzag.Encode(int64(x))) }
func EncodeVarInt16(x int16) []byte { return EncodeVarInt64U(zigzag.Encode(int64(x))) }

func DecodeVarInt(data []byte) (x int64, nRead int, smallOrOverflow, ok bool) {
	// var zz uint64
	x, nRead = binary.Varint(data)
	if nRead == 0 {
		smallOrOverflow = true
	} else if nRead > 0 {
		ok = true
		// x = zigzag.Decode(zz)
	} else {
		nRead = -nRead
	}
	return
}

func EncodeFloat32(f float32) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

func DecodeFloat32(data []byte) (ret float32, n int, err error) {
	r := bytes.NewReader(data)
	if err = binary.Read(r, ByteOrder, &ret); err == nil {
		var pos int64
		pos, err = r.Seek(0, io.SeekCurrent)
		n = int(pos)
	}
	return
}

func float64ToByte(f float64) []byte {
	var buf [8]byte
	ByteOrder.PutUint64(buf[:], math.Float64bits(f))
	return buf[:]
}

func EncodeFloat64(f float64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

func DecodeFloat64(data []byte) (ret float64, n int, err error) {
	r := bytes.NewReader(data)
	if err = binary.Read(r, ByteOrder, &ret); err == nil {
		var pos int64
		pos, err = r.Seek(0, io.SeekCurrent)
		n = int(pos)
	}
	return
}

func EncodeFixed16To(data []byte, pos int, f int16) {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	copy(data[pos:], buf.Bytes())
}

func EncodeFixed32To(data []byte, pos int, f int32) {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	copy(data[pos:], buf.Bytes())
}

func EncodeFixed32(f int32) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

func EncodeFixed64(f int64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

func DecodeFixed64(data []byte) (ret int64, n int, err error) {
	r := bytes.NewReader(data)
	if err = binary.Read(r, ByteOrder, &ret); err == nil {
		var pos int64
		pos, err = r.Seek(0, io.SeekCurrent)
		n = int(pos)
	}
	return
}

func EncodeFixedU64(f uint64) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, ByteOrder, f)
	if err != nil {
		fmt.Println("      binary.Write failed: ", err)
	}
	return buf.Bytes()
}

func DecodeFixedU64(data []byte) (ret float32, n int, err error) {
	r := bytes.NewReader(data)
	if err = binary.Read(r, ByteOrder, &ret); err == nil {
		var pos int64
		pos, err = r.Seek(0, io.SeekCurrent)
		n = int(pos)
	}
	return
}
