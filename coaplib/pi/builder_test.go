package pi_test

import (
	"github.com/hedzr/go-socketlib/coaplib/message"
	"github.com/hedzr/go-socketlib/coaplib/pi"
	"net/url"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	builder := pi.NewBuilder()
	var msg *message.Message

	builder.Reset().
		WithURI2(url.Parse("coap://coap.me/test"))
	msg = builder.Build()
	if builder.Error() != nil {
		t.Fatal(builder.Error())
	}
	t.Logf("message: %v", msg)
	t.Logf("message: % x", msg.AsBytes())

	builder.NewBase("coap://coap.me").
		WithURIPath("/test")
	msg = builder.Build()
	if builder.Error() != nil {
		t.Fatal(builder.Error())
	}
	t.Logf("message: %v", msg)
	t.Logf("message: % x", msg.AsBytes())
}
