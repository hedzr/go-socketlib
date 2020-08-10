package message

import (
	"bytes"
	"math"
	"testing"
)

func TestRoundedLog2(t *testing.T) {
	for _, st := range []struct {
		v   uint
		exp byte
	}{
		{16, 0},
		{32, 1},
		{64, 2},
		{128, 3},
		{256, 4},
		{512, 5},
		{1024, 6},
	} {
		r := RoundedLog2(st.v)
		if math.Pow(2, float64(r)) != float64(st.v) && r-4 == st.exp {
			t.Fatalf("failed: log2(%v)=%v | but got %v", st.v, st.exp, r)
		}
	}
}

func TestBuilder_block1Encode(t *testing.T) {
	for i := 0; i < 32; i++ {
		hi, lo := bigEndianHalfEncodeUint64(uint64(i))
		t.Logf("%v => %v,%02X", i, hi, lo<<4)
	}
}

func TestOptHeaderEncodeDeltaAndLength(t *testing.T) {
	var bb bytes.Buffer
	var oneByteBuffer = []byte{0}

	delta, length := OptionNumberURIHost, 23

	o := optBase{}

	_ = o.WriteOptDeltaAndLength(&bb, oneByteBuffer, delta, length)
}

func TestLogger(t *testing.T) {

}
