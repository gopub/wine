package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

func Bytes(status int, b []byte) *Response {
	typ := http.DetectContentType(b)
	header := make(http.Header)
	header.Set(mime.ContentType, typ)
	return &Response{
		status: status,
		header: header,
		value:  b,
	}
}
