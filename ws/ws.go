package ws

import (
	"net"
	"time"

	"github.com/gopub/types"

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

	// server side
	remoteAddr net.Addr

	// client side
	createdAt time.Time
}

func (r *Request) RemoteAddr() net.Addr {
	return r.remoteAddr
}

type Response struct {
	ID    int64         `json:"id,omitempty"`
	Data  interface{}   `json:"data,omitempty"`
	Error *errors.Error `json:"error,omitempty"`
}

func (r *Response) IsPush() bool {
	return r.ID != 0 && r.ID%2 == 0
}

const (
	methodGetDate = "ws.getDate"
)
