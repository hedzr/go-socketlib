package pi

import (
	"reflect"
	"testing"
)

func TestLeadBytesS_OnEncode(t *testing.T) {
	co := NewLeadBytes([]byte{0x55, 0xaa})
	for i, c := range []struct{ in, out []byte }{
		{[]byte{1, 2, 3}, []byte{0x55, 0xaa, 6, 1, 2, 3}},
	} {
		t.Logf("----------------- %5d. input: %v", i, c.in)
		out, err := co.OnEncode(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out, c.out) {
			t.Fatalf("  expect encode output is %v, but got %v", c.out, out)
		}

		ch := make(chan []byte, 8)
		processed, err := co.OnDecode(out, ch)
		if err != nil {
			t.Fatal(err)
		}
		if !processed {
			t.Fatal("not processed")
		}
		in := <-ch
		close(ch)
		if !reflect.DeepEqual(in, c.in) {
			t.Fatalf("  expect decode output is %v, but got %v", c.in, in)
		}
	}
}

func TestLeadBytesS_OnDecode(t *testing.T) {
}
