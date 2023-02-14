package codec

import (
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
)

// This could use any io.Reader as input, for example
// a request body in http requests
func decoder(r io.Reader) io.Reader {

	// We simply set up a custom chain of Decoders
	d, err := zlib.NewReader(
		base64.NewDecoder(base64.StdEncoding, r))

	// This should only occur if one of the Decoders can not reset
	// its internal buffer. Hence, it validates a panic.
	if err != nil {
		panic(fmt.Sprintf("Error setting up decoder chain: %s", err))
	}

	// We return an io.Reader which can be used as any other
	return d

}
