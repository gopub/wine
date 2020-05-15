package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

// HTML creates a HTML response
func HTML(status int, html string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.HtmlUTF8)
	return &Response{
		status: status,
		header: header,
		value:  html,
	}
}
