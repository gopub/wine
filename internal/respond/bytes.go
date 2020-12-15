package respond

import (
	"net/http"

	"github.com/gopub/wine/httpvalue"
)

func Bytes(status int, b []byte) *Response {
	typ := http.DetectContentType(b)
	header := make(http.Header)
	header.Set(httpvalue.ContentType, typ)
	return &Response{
		status: status,
		header: header,
		value:  b,
	}
}
