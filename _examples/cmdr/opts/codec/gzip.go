package codec

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func decompressByGzip(data []byte) (result []byte, nRead int, err error) {
	var b = bytes.NewReader(data)
	var gz *gzip.Reader
	gz, err = gzip.NewReader(b)
	if nRead, err = gz.Read(data); err == nil {
		result, err = ioutil.ReadAll(gz)
		if err == nil {
			err = gz.Close()
		} else {
			_ = gz.Close()
		}
	}
	return
}

func compressByGzip(data []byte) (result []byte, err error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err = gz.Write(data); err == nil {
		if err = gz.Close(); err == nil {
			result = b.Bytes()
		}
	}
	return
}
