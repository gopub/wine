package respond

import (
	"net/http"

	"github.com/gopub/wine/mime"
)

func Redirect(location string, permanent bool) *Response {
	header := make(http.Header)
	header.Set("Location", location)
	header.Set(mime.ContentType, mime.Plain)
	var status int
	if permanent {
		status = http.StatusMovedPermanently
	} else {
		status = http.StatusFound
	}

	return &Response{
		status: status,
		header: header,
	}
}
