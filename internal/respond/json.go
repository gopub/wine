package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

// JSON creates a application/json response
func JSON(status int, value interface{}) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.JsonUTF8)
	return &Response{
		status: status,
		header: header,
		value:  value,
	}
}
