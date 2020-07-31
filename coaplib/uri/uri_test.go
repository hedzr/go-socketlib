package uri_test

import (
	uri2 "github.com/hedzr/go-socketlib/coaplib/uri"
	"testing"
)

func TestURIs(t *testing.T) {
	for i, d := range []struct {
		uri string
	}{
		{"coap://"},
		{"coaps://"},
		{"coap://example.com:5683/~serial/test"},
		{"coap://EXAMPLE.COM/%7eserial/test"},
		{"coap://example.com:5683/~serial/test?"},
		{"coap://example.com:5683/~serial/test?q="},
		{"coap://example.com:5683/~serial/test?q=best"},
		{"coap://host.com/temp/txt"},
		{"coap://host/temp/txt"},
		{"coap://host:10000/temp/txt"},
		{"coaps://host/temp/txt"},
		{"coaps://host:10000/temp/txt"},
		{"coap://user:pass@host/temp/txt"},
		{"coap://user:pass@host:10000/temp/txt"},
		{"coaps://user:pass@host/temp/txt"},
		{"coaps://user:pass@host:10000/temp/txt"},
	} {
		uri, err := uri2.ParseURI(d.uri)
		if err != nil {
			t.Fatalf("%5d. %50q -> url failed: %v", i, d.uri, err)
		} else {
			t.Logf("%5d. %50q -> url OK: %v / port: %v", i, d.uri, uri, uri.Port())
		}
	}
}
