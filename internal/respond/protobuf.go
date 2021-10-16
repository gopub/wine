package respond

import (
	"net/http"

	"github.com/gopub/wine/httpvalue"
	"google.golang.org/protobuf/proto"
)

func Protobuf(status int, message proto.Message) *Response {
	header := make(http.Header)
	header.Set(httpvalue.ContentType, httpvalue.Protobuf)
	return &Response{
		status: status,
		header: header,
		value:  message,
	}
}
