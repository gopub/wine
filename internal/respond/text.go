package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

// Status returns a response only with a status code
func Status(status int) *Response {
	return PlainText(status, http.StatusText(status))
}

// PlainText sends a text/plain response
func PlainText(status int, text string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.PlainUTF8)
	if text == "" {
		text = http.StatusText(status)
	}
	return &Response{
		status: status,
		header: header,
		value:  text,
	}
}

// CSS sends a text/css response
func CSS(status int, css string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.CSS)
	if css == "" {
		css = http.StatusText(status)
	}
	return &Response{
		status: status,
		header: header,
		value:  css,
	}
}
