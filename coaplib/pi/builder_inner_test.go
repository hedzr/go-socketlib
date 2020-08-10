package pi

import (
	"encoding/hex"
	"github.com/hedzr/go-socketlib/coaplib/message"
	"net/url"
	"testing"
)

func TestBuilder_block1(t *testing.T) {

	builder := NewBuilder()
	var msg *message.Message
	// var err error

	//const block13 = "4801b40184cf580326c20da73d0a63616c69666f726e69756d2e65636c697073652e6f7267856c6172676560620112"
	//block13bytes, err := hex.DecodeString(block13)

	const block17 = "4801b40184cf580326c20da73d0a63616c69666f726e69756d2e65636c697073652e6f7267856c6172676560620112"
	block17bytes, err := hex.DecodeString(block17)

	if err != nil {
		t.Fatal(err, block17bytes)
	}

	// data := nil // block17bytes[0:]
	block1 := message.NewBlock1(13, false, 64, nil)
	builder.
		WithMessageID(1).
		WithToken(0x123456789abcdef0).
		WithURI2(url.Parse("coap://californium.eclipse.org/large")).
		WithAccept(message.TextPlain).
		WithBlock1Append(block1)
	msg = builder.Build()

	if builder.Error() != nil {
		t.Fatalf("build failed: %v", builder.Error())
	}

	t.Logf("msg: % x", msg.AsBytes())
}
