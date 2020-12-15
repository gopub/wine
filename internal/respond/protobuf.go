package respond

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/wine/httpvalue"
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
