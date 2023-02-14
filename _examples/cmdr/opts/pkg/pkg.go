package pkg

import (
	"bytes"
	"github.com/hedzr/go-socketlib/_examples/cmdr/opts/codec"
)

type (
	Builder interface {
		New() Builder
		Build() Package
	}

	Package interface {
		Bytes(reservedBytes int) []byte
	}

	PackageBuilder interface {
		Builder
		Command(cmd Command) PackageBuilder
		Body(body []byte) PackageBuilder
		SendOOB(oob []byte) PackageBuilder
	}
)

type (
	Pkg struct {
		Command Command
		ObjID   string
		Body    []byte
	}

	Command int16
)

const (
	PCDummy   Command = iota
	PCOOb     Command = 1
	PCSeqData Command = 2
	PCBool    Command = 3
	PCFloat   Command = 4
)

func (pkg *Pkg) Bytes(reservedBytes int) []byte {
	return ToBytes(pkg, reservedBytes)
}

func ToBytes(pkg *Pkg, reservedBytes int) []byte {
	var sb bytes.Buffer

	// the reserved leading bytes
	sb.Write(bytes.Repeat([]byte{0}, reservedBytes))

	b := codec.EncodeVarInt16(int16(pkg.Command))
	sb.Write(b)

	b = codec.EncodeString(pkg.ObjID)
	sb.Write(b)

	sb.Write(pkg.Body)

	return sb.Bytes()
}
