package zigzag

func Encode(x int64) uint64 {
	// 左移一位 XOR (-1 / 0 的 64 位补码)
	return (uint64(x) << 1) ^ uint64(x>>63)
}

func Decode(zz uint64) int64 {
	return int64(zz>>1) ^ -int64(zz&1)
}
