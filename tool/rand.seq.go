/*
 * Copyright © 2020 Hedzr Yeh.
 */

/*
 */

package tool

import (
	"math/rand"
	"strconv"
	"time"
)

// Intv 忽略任何错误转换字符串为整数值并返回。如果无法转换，返回值为0
func Intv(s string) (v int) {
	v, _ = strconv.Atoi(s)
	return
}

// RandSeq returns a random string with specified length 'n'
func RandSeq(n int) string {
	b := make([]rune, n)
	cs := []rune(charset)
	for i := range b {
		b[i] = cs[seededRand.Intn(len(cs))]
	}
	return string(b)
}

// RandSeqln returns a random string with specified length 'n', and it has a '\n' character tail.
func RandSeqln(n int) string {
	b := make([]rune, n+1)
	cs := []rune(charset)
	for i := range b {
		b[i] = cs[seededRand.Intn(len(cs))]
	}
	b[n] = '\n'
	return string(b)
}

// RandRandSeq returns a random string with random length (1..127)
func RandRandSeq() string {
	n := seededRand.Intn(128)
	return RandSeq(n)
}

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// StringWithCharset generate random string
func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// String returns a random string with length specified
func String(length int) string {
	return StringWithCharset(length, charset)
}

// String returns a random string with length specified
func StringVariantLength(minLength, maxLength int) string {
	if minLength <= 0 {
		minLength = 1
	}
	if maxLength <= minLength+1 {
		maxLength = minLength + 4096
	}
	length := seededRand.Intn(maxLength-minLength) + minLength
	return StringWithCharset(length, charset)
}
