package codec

import (
	"time"
)

func EncodeTimestamp(ts *time.Time) (r []byte) {
	r = EncodeFixed64(ts.UnixNano())
	return
}

func DecodeTimestamp(data []byte) (ts time.Time, n int, err error) {
	var nano int64
	nano, n, err = DecodeFixed64(data)
	if err == nil {
		ts = time.Unix(0, nano)
	}
	return
}

func EncodeTimestamp9bytes(ts *time.Time) (r []byte) {
	r = EncodeVarInt64U(uint64(ts.UnixNano()))
	return
}

func DecodeTimestamp9bytes(data []byte) (ts time.Time, n int, ok bool) {
	var nano uint64
	nano, n, _, ok = DecodeVarIntU(data)
	if ok {
		ts = time.Unix(0, int64(nano))
	}
	return
}

func TimestampToString(t time.Time) string { return formatTimestamp(t) }
func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func EncodeTimestamp15bytes(ts *time.Time) (r []byte, err error) {
	r, err = ts.MarshalBinary()
	// if err != nil {
	//	log.Errorf("cannot time.Time::MarshalBinary() on %v: %v", ts, err)
	// }
	// return r
	return
}

func DecodeTimestamp15bytes(data []byte) (ts *time.Time, err error) {
	ts = new(time.Time)
	// if err = ts.UnmarshalBinary(data); err != nil {
	//	log.Errorf("cannot time.Time::UnmarshalBinary(%v): %v", data, err)
	//	ts = nil
	// }
	err = ts.UnmarshalBinary(data)
	return
}
