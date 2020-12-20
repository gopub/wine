package respond

import (
	"net/http"

	"github.com/gopub/wine/httpvalue"
)

func Redirect(location string, permanent bool) *Response {
	header := make(http.Header)
	header.Set(httpvalue.Location, location)
	header.Set(httpvalue.ContentType, httpvalue.Plain)
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
