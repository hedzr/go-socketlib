package codec

func EncodeBool(ts bool) []byte {
	r := make([]byte, 1)
	if ts {
		r[0] = 1
	} else {
		r[0] = 0
	}
	return r
}

func DecodeBool(data []byte) bool {
	if len(data) > 0 {
		if data[0] == 1 {
			return true
		}
	}
	return false
}
