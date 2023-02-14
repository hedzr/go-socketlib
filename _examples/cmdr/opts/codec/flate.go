package codec

import (
	"bytes"
	"compress/flate"
	"io"
)

func decompressByDeflate(data []byte) (result []byte, nWritten int, err error) {
	r := flate.NewReader(bytes.NewReader(data))
	var bw bytes.Buffer
	var written int64
	written, err = io.Copy(&bw, r)
	if err == nil {
		result, nWritten, err = bw.Bytes(), int(written), r.Close()
	} else {
		_ = r.Close()
	}
	return
}

func compressByFlate(data []byte) (result []byte, err error) {
	var b bytes.Buffer
	var w *flate.Writer
	if w, err = flate.NewWriter(&b, flate.DefaultCompression); err != nil {
		return
	}
	if _, err = w.Write(data); err == nil {
		//if _, err := io.Copy(w, strings.NewReader(data)); err == nil {}
		if err = w.Close(); err == nil {
			result = b.Bytes()
		}
	}
	return
}
