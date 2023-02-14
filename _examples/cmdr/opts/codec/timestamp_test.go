package codec

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestTimestampFormat(tt *testing.T) {
	// Parse a time value from a string in the standard Unix format.
	t, err := time.Parse(time.UnixDate, "Wed Feb 25 11:06:39 PST 2015")
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}

	// time.Time's Stringer method is useful without any format.
	fmt.Println("default format:", t)

	// Predefined constants in the package implement common layouts.
	fmt.Println("Unix format:", t.Format(time.UnixDate))

	// The time zone attached to the time value affects its output.
	fmt.Println("Same, in UTC:", t.UTC().Format(time.UnixDate))

	// The rest of this function demonstrates the properties of the
	// layout string used in the format.

	// The layout string used by the Parse function and Format method
	// shows by example how the reference time should be represented.
	// We stress that one must show how the reference time is formatted,
	// not a time of the user's choosing. Thus each layout string is a
	// representation of the time stamp,
	//	Jan 2 15:04:05 2006 MST
	// An easy way to remember this value is that it holds, when presented
	// in this order, the values (lined up with the elements above):
	//	  1 2  3  4  5    6  -7
	// There are some wrinkles illustrated below.

	// Most uses of Format and Parse use constant layout strings such as
	// the ones defined in this package, but the interface is flexible,
	// as these examples show.

	// Define a helper function to make the examples' output look nice.
	do := func(name, layout, want string) {
		got := t.Format(layout)
		if want != got {
			fmt.Printf("error: for %q got %q; expected %q\n", layout, got, want)
			return
		}
		fmt.Printf("%-16s %q gives %q\n", name, layout, got)
	}

	// Print a header in our output.
	fmt.Printf("\nFormats:\n\n")

	// Simple starter examples.
	do("Basic full date", "Mon Jan 2 15:04:05 MST 2006", "Wed Feb 25 11:06:39 PST 2015")
	do("Basic short date", "2006/01/02", "2015/02/25")

	// The hour of the reference time is 15, or 3PM. The layout can express
	// it either way, and since our value is the morning we should see it as
	// an AM time. We show both in one format string. Lower case too.
	do("AM/PM", "3PM==3pm==15h", "11AM==11am==11h")

	// When parsing, if the seconds value is followed by a decimal point
	// and some digits, that is taken as a fraction of a second even if
	// the layout string does not represent the fractional second.
	// Here we add a fractional second to our time value used above.
	t, err = time.Parse(time.UnixDate, "Wed Feb 25 11:06:39.1234 PST 2015")
	if err != nil {
		panic(err)
	}
	// It does not appear in the output if the layout string does not contain
	// a representation of the fractional second.
	do("No fraction", time.UnixDate, "Wed Feb 25 11:06:39 PST 2015")

	// Fractional seconds can be printed by adding a run of 0s or 9s after
	// a decimal point in the seconds value in the layout string.
	// If the layout digits are 0s, the fractional second is of the specified
	// width. Note that the output has a trailing zero.
	do("0s for fraction", "15:04:05.00000", "11:06:39.12340")

	// If the fraction in the layout is 9s, trailing zeros are dropped.
	do("9s for fraction", "15:04:05.99999999", "11:06:39.1234")

}

func TestTimestampEncodeDecode(t *testing.T) {
	for _, c := range []struct {
		src time.Time
		res []byte
	}{
		{time.Date(2009, time.November, 10, 23, 51, 37, 1379, time.UTC),
			[]byte{17, 116, 242, 190, 190, 174, 31, 99}},
	} {
		buf := EncodeTimestamp(&c.src)
		t.Logf("src=%v => count=%v, buf=%v", formatTimestamp(c.src), len(buf), buf)
		if !reflect.DeepEqual(buf, c.res) {
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		tgt, n, err := DecodeTimestamp(buf)
		if err != nil {
			t.Errorf("wrong decode timestamp from %v", buf)
		} else {
			t.Logf("decoded: tgt=%v, %v bytes read.", formatTimestamp(tgt), n)
			if tgt.UnixNano() != c.src.UnixNano() {
				t.Errorf("expecting %v but got %v", formatTimestamp(c.src), formatTimestamp(tgt))
			}
		}
	}
}

func TestTimestamp9bytesEncodeDecode(t *testing.T) {
	for _, c := range []struct {
		src time.Time
		res []byte
	}{
		{time.Date(2009, time.November, 10, 23, 51, 37, 1379, time.UTC),
			[]byte{227, 190, 184, 245, 235, 215, 188, 186, 17}},
	} {
		buf := EncodeTimestamp9bytes(&c.src)
		t.Logf("src=%v => count=%v, buf=%v", formatTimestamp(c.src), len(buf), buf)
		if !reflect.DeepEqual(buf, c.res) {
			t.Errorf("expecting %v but got %v", c.res, buf)
		}

		tgt, n, ok := DecodeTimestamp9bytes(buf)
		if !ok {
			t.Errorf("wrong decode timestamp from %v", buf)
		} else {
			t.Logf("decoded: tgt=%v, %v bytes read.", formatTimestamp(tgt), n)
			if tgt.UnixNano() != c.src.UnixNano() {
				t.Errorf("expecting %v but got %v", formatTimestamp(c.src), formatTimestamp(tgt))
			}
		}
	}
}

func TestTimestamp15bytesEncodeDecode(t *testing.T) {
	for _, c := range []struct {
		src time.Time
		res []byte
	}{
		{time.Date(2009, time.November, 10, 23, 51, 37, 1379, time.UTC),
			[]byte{1, 0, 0, 0, 14, 194, 139, 243, 137, 0, 0, 5, 99, 255, 255}},
	} {
		buf, err := EncodeTimestamp15bytes(&c.src)
		if err != nil {
			t.Errorf("wrong encode timestamp from %v: err = %v", formatTimestamp(c.src), err)
		} else {
			t.Logf("src=%v => count=%v, buf=%v", formatTimestamp(c.src), len(buf), buf)
			if !reflect.DeepEqual(buf, c.res) {
				t.Errorf("expecting %v but got %v", c.res, buf)
			}
		}

		var tgt *time.Time
		tgt, err = DecodeTimestamp15bytes(buf)
		if err != nil {
			t.Errorf("wrong decode timestamp from %v: err = %v", buf, err)
		} else {
			t.Logf("decoded: tgt=%v.", formatTimestamp(*tgt))
			if tgt.UnixNano() != c.src.UnixNano() {
				t.Errorf("expecting %v but got %v", formatTimestamp(c.src), formatTimestamp(*tgt))
			}
		}
	}
}
