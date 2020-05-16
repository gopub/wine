package respond

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/wine/mime"
)

func Protobuf(status int, message proto.Message) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.Protobuf)
	return &Response{
		status: status,
		header: header,
		value:  message,
	}
}
