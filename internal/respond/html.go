package respond

import (
	"net/http"

	"github.com/gopub/wine/httpvalue"
)

// HTML creates a HTML response
func HTML(status int, html string) *Response {
	header := make(http.Header)
	header.Set(httpvalue.ContentType, httpvalue.HtmlUTF8)
	return &Response{
		status: status,
		header: header,
		value:  html,
	}
}
