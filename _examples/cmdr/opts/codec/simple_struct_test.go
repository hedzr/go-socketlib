package codec

import (
	"reflect"
	"testing"
)

type ts struct {
	Pi  float64
	I64 int64
	// Str string
}

func TestStructEncodeDecode(t *testing.T) {
	for _, c := range []struct {
		src ts
		res []byte
	}{
		{ts{3.141592653589794, -1},
			[]byte{64, 9, 33, 251, 84, 68, 45, 26, 255, 255, 255, 255, 255, 255, 255, 255}},
	} {
		buf := EncodeStruct(&c.src)
		t.Logf("src=%v => count=%v, buf=%v", c.src, len(buf), buf)
		if !reflect.DeepEqual(buf, c.res) {
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		var tgt = new(ts)
		nRead, err := DecodeStruct(buf, tgt)
		if err != nil {
			t.Errorf("wrong decode num from %v: err = %v", buf, err)
		} else {
			t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, nRead)
			if !reflect.DeepEqual(tgt, &c.src) {
				t.Errorf("expecting %v but got %v", c.src, tgt)
			}
		}
	}
}
