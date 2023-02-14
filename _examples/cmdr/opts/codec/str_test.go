package codec

import (
	"reflect"
	"testing"
)

func TestFlateEncodeDecode(t *testing.T) {
	data := []byte("hello")
	res, err := compressByFlate(data)
	if err == nil {
		de, n, e := decompressByDeflate(res)
		err = e
		t.Logf("de [%v]: %v", n, de)
		if !reflect.DeepEqual(data, de) {
			t.Errorf("want %v but got %v", data, de)
		}
	}
	if err != nil {
		t.Error(err)
	}
}

func TestStringEncodeDecode(t *testing.T) {
	for _, c := range []struct {
		src string
		res []byte
	}{
		{"hello", []byte{9, 104, 101, 108, 108, 111}},
		{"world", []byte{9, 119, 111, 114, 108, 100}},
		{"Be aware that they may be removed or changed in future versions of fish.\n\n",
			[]byte{142, 1, 4, 192, 193, 17, 196, 32, 8, 5, 208, 187, 85,
				252, 10, 182, 136, 237, 132, 196, 111, 240, 160, 204, 0, 154,
				177, 251, 188, 63, 33, 175, 56, 145, 42, 137, 84, 30, 12, 57,
				184, 8, 231, 176, 205, 10, 115, 220, 42, 243, 97, 69, 159, 104,
				43, 151, 19, 155, 30, 221, 102, 192, 26, 90, 15, 253, 149, 242,
				5, 0, 0, 255, 255}},
		{"For examples of how to write your own complex completions, study the completions in /usr/share/fish/completions.",
			[]byte{170, 1, 76, 138, 81, 10, 2, 49, 12, 5, 175, 242, 14, 32,
				230, 22, 222, 99, 209, 44, 9, 104, 159, 36, 41, 109, 111, 47,
				194, 126, 244, 107, 96, 102, 30, 12, 232, 60, 62, 223, 183, 38,
				120, 194, 56, 80, 196, 8, 47, 197, 98, 15, 112, 52, 60, 249,
				31, 230, 197, 114, 182, 188, 33, 171, 191, 22, 202, 116, 215,
				240, 6, 233, 25, 146, 118, 132, 202, 233, 105, 178, 229, 251,
				47, 0, 0, 255, 255}},
		{"hello hello hello hello hello hello hello ",
			[]byte{28, 202, 72, 205, 201, 201, 87, 32, 134, 4, 4, 0, 0, 255, 255}},
	} {
		buf := EncodeString(c.src)
		t.Logf("src: %v => buf: %v", c.src, buf)
		t.Logf("     %v bytes => %v bytes, %v", len(c.src), len(buf), float32(len(buf))/float32(len(c.src)))
		if !reflect.DeepEqual(buf, c.res) {
			t.Errorf("Error: expecting %v but got %v", c.res, buf)
		}

		tgt, nRead, err := DecodeString(buf)
		if err != nil {
			t.Errorf("wrong decode num from %v: err = %v", buf, err)
		} else {
			t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, nRead)
			if tgt != c.src {
				t.Errorf("Error: expecting %v but got %v", c.src, tgt)
			}
		}
	}
}
