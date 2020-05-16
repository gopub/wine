package rpc

import (
	"net/http"

	"github.com/gopub/types"
	"github.com/gopub/wine"
)

type coder interface {
	Code() int
}

type messageCoder interface {
	coder
	Message() string
}

type Result struct {
	Error *types.Error `json:"error,omitempty"`
	Data  interface{}  `json:"data"`
}

func Data(data interface{}) wine.Responder {
	val := &Result{
		Data: data,
	}
	return wine.JSON(http.StatusOK, val)
}

func StatusData(status int, data interface{}) wine.Responder {
	val := &Result{
		Data: data,
	}
	return wine.JSON(status, val)
}
