package respond

import (
	"net/http"

	"github.com/gopub/wine/httpvalue"
)

// JSON creates a application/json response
func JSON(status int, value interface{}) *Response {
	header := make(http.Header)
	header.Set(httpvalue.ContentType, httpvalue.JsonUTF8)
	return &Response{
		status: status,
		header: header,
		value:  value,
	}
}
