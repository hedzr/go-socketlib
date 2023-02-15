package codec

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/codec/zigzag"
)

func TestVarintU(t *testing.T) {
	// var src uint64 = 789
	buf := make([]byte, 11)
	var tgt uint64

	for _, c := range []struct {
		src uint64
		res []byte
	}{
		{1, []byte{1}},
		{2, []byte{2}},
		{3, []byte{3}},
		{4, []byte{4}},
		{5, []byte{5}},
		{6, []byte{6}},
		{127, []byte{127}},
		{128, []byte{128, 1}},
		{255, []byte{255, 1}},
		{256, []byte{128, 2}},
		{789, []byte{149, 6}},
		{789123654, []byte{198, 164, 164, 248, 2}},
	} {
		count := binary.PutUvarint(buf, c.src)
		if !reflect.DeepEqual(buf[:len(c.res)], c.res) {
			t.Logf("src=%v => count=%v, buf=%v", c.src, count, buf)
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		tgt, count = binary.Uvarint(buf)
		if tgt != c.src {
			t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, count)
			t.Errorf("expecting %v but got %v", c.src, tgt)
		}
	}
}

func TestVarint(t *testing.T) {
	// var src uint64 = 789
	buf := make([]byte, 11)
	var tgt int64

	for _, c := range []struct {
		src int64
		res []byte
	}{
		{1, []byte{2}},
		{2, []byte{4}},
		{3, []byte{6}},
		{4, []byte{8}},
		{5, []byte{10}},
		{6, []byte{12}},
		{127, []byte{254, 1}},
		{128, []byte{128, 2}},
		{255, []byte{254, 3}},
		{256, []byte{128, 4}},
		{789, []byte{170, 12}},
		{789123654, []byte{140, 201, 200, 240, 5}},
		// {-1, []byte{1}},
	} {
		count := binary.PutVarint(buf, c.src)
		if !reflect.DeepEqual(buf[:count], c.res) {
			t.Logf("src=%v => count=%v, buf=%v", c.src, count, buf)
			t.Errorf("Error: expecting %v but got %v", c.res, buf)
		}

		tgt, count = binary.Varint(buf)
		if tgt != c.src {
			t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, count)
			t.Errorf("Error: expecting %v but got %v", c.src, tgt)
		}
	}
}

func TestZigZaggedVarint(t *testing.T) {
	// var src uint64 = 789
	buf := make([]byte, 11)
	var tgt uint64

	for _, c := range []struct {
		src int64
		res []byte
	}{
		{1, []byte{2}},
		{2, []byte{4}},
		{3, []byte{6}},
		// {4, []byte{4}},
		// {5, []byte{5}},
		// {6, []byte{6}},
		// {127, []byte{127}},
		// {128, []byte{128, 1}},
		// {255, []byte{255, 1}},
		// {256, []byte{128, 2}},
		// {789, []byte{149, 6}},
		// {789123654, []byte{198, 164, 164, 248, 2}},
		{-1, []byte{1}},
		{-2, []byte{3}},
		{-3, []byte{5}},
	} {
		num := zigzag.Encode(c.src)
		count := binary.PutUvarint(buf, num)
		t.Logf("src=%v => count=%v, buf=%v", c.src, count, buf)
		if !reflect.DeepEqual(buf[:len(c.res)], c.res) {
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		tgt, count = binary.Uvarint(buf)
		t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, count)
		if tgt != num {
			t.Errorf("expecting %v but got %v", num, tgt)
		}
		tnt := zigzag.Decode(tgt)
		if tnt != c.src {
			t.Errorf("expecting %v but got %v", c.src, tnt)
		}
	}
}

func TestNumEncodeDecode(t *testing.T) {
	// var src uint64 = 789
	// buf := make([]byte, 11)
	// var tgt uint64

	for _, c := range []struct {
		src int64
		res []byte
	}{
		{1, []byte{2}},
		{2, []byte{4}},
		{3, []byte{6}},
		// {4, []byte{4}},
		// {5, []byte{5}},
		// {6, []byte{6}},
		// {127, []byte{127}},
		// {128, []byte{128, 1}},
		// {255, []byte{255, 1}},
		// {256, []byte{128, 2}},
		// {789, []byte{149, 6}},
		// {789123654, []byte{198, 164, 164, 248, 2}},
		{-1, []byte{1}},
		{-2, []byte{3}},
		{-3, []byte{5}},
		{2147483647, []byte{254, 255, 255, 255, 15}},
		{-2147483647, []byte{253, 255, 255, 255, 15}},
	} {
		buf := EncodeVarInt(c.src)
		t.Logf("src=%v => count=%v, buf=%v", c.src, len(buf), buf)
		if !reflect.DeepEqual(buf, c.res) {
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		tgt, nRead, smallOrOverflow, ok := DecodeVarInt(buf)
		if !ok {
			t.Errorf("wrong decode num from %v: smallOrOverflow = %v", buf, smallOrOverflow)
		} else {
			t.Logf("decoded: tgt=%v, %v bytes ate.", tgt, nRead)
			if tgt != c.src {
				t.Errorf("expecting %v but got %v", c.src, tgt)
			}
		}
	}
}
