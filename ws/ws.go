package ws

import (
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
