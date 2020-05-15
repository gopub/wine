package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

// Status returns a response only with a status code
func Status(status int) *Response {
	return Text(status, http.StatusText(status))
}

// Text sends a text response
func Text(status int, text string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.PlainUTF8)
	return &Response{
		status: status,
		header: header,
		value:  text,
	}
}
