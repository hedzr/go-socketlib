package opts

import (
	"bytes"
	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/codec"
	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/pkg"
	"github.com/hedzr/log"
	"sync/atomic"
)

func buildPkg(i, threadId int, destAddr string) (data []byte) {
	data = putBoolWith(i)
	putLeadingBytes(data)
	atomic.AddInt64(&totalBuilt, 1)
	log.Debugf("%7d: package (#%v) built: %v", i, totalBuilt, data)
	return
}

func putBoolWith(i int) (data []byte) {
	state := true
	if i/2*2 == i {
		state = false
	}

	var body bytes.Buffer
	bl := codec.EncodeBool(state)
	body.Write(bl)
	bl = codec.EncodeVarInt64U(uint64(totalBuilt))
	body.Write(bl)

	b := pkg.NewBuilder().(pkg.PackageBuilder)
	b.Command(pkg.PCBool).Body(body.Bytes())
	data = b.Build().Bytes(reservedBytes)
	return
}

func putLeadingBytes(data []byte) {
	// the reserved leading bytes
	data[0] = 0xaa
	data[1] = 0x55
	l := len(data) - reservedBytes
	codec.EncodeFixed16To(data, 2, int16(l))
}

var totalBuilt int64

const reservedBytes = 4 // 0xaa,0x55, length (fixed int16)
