package ws

import (
	"github.com/gopub/types"
	"net"

	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/wine"
)

var logger = wine.Logger()

func SetLogger(l *log.Logger) {
	logger = l
}

type ReadWriter interface {
	ReadJSON(i interface{}) error
	WriteJSON(i interface{}) error
}

type GetAuthUserID interface {
	GetAuthUserID() int64
}

type Request struct {
	ID     int64       `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Header types.M     `json:"header,omitempty"`
	Body   interface{} `json:"body,omitempty"`

	remoteAddr net.Addr
}

func (r *Request) RemoteAddr() net.Addr {
	return r.remoteAddr
}

type Response struct {
	ID    int64         `json:"id,omitempty"`
	Data  interface{}   `json:"data,omitempty"`
	Error *errors.Error `json:"error,omitempty"`
}
