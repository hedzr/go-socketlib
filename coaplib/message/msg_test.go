package message

import (
	"encoding/hex"
	"testing"
)

func TestCode_String(t *testing.T) {
	for i := Code(0); i < Code(0xff); i++ {
		t.Log(i, i.Category(), i.Details(), int(i))
	}
	t.Log(Code(0xff))
}

func TestMessage_Decode(t *testing.T) {
	data, err := hex.DecodeString("480109d6fc911b24d83d851637636f61702e6d65847061746860")

	var m Message

	t.Logf("data: % x", data)
	err = m.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("decoded: %v", m.String())
}
