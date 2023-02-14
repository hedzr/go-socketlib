package zigzag

import "testing"

func TestZigZag64(t *testing.T) {
	for _, c := range []struct {
		src int64
		res uint64
	}{
		{0, 0},
		{-1, 1},
		{1, 2},
		{2147483647, 4294967294},
		{-2147483648, 4294967295},
	} {
		num := Encode(c.src)
		if num != c.res {
			t.Errorf("expecting %v but got %v", c.res, num)
		}
	}
}

func TestZigZag64Decode(t *testing.T) {
	for _, c := range []struct {
		res int64
		src uint64
	}{
		{0, 0},
		{-1, 1},
		{1, 2},
		{2147483647, 4294967294},
		{-2147483648, 4294967295},
	} {
		num := Decode(c.src)
		if num != c.res {
			t.Errorf("expecting %v but got %v", c.res, num)
		}
	}
}
