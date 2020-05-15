package rpc

import (
	"net/http"

	"github.com/gopub/types"
	"github.com/gopub/wine"
)

const (
	StatusTransportFailed = 600
	StatusInvalidResponse = 601
)

func Errorf(code int, msgFormat string, msgArgs ...interface{}) wine.Responder {
	val := &Result{
		Error: types.NewError(code, msgFormat, msgArgs...),
	}
	status := code
	for status >= 1000 {
		status /= 10
	}
	return wine.JSON(status, val)
}

func Error(err error) wine.Responder {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	if e, ok := err.(messageCoder); ok {
		return Errorf(e.Code(), e.Message())
	} else if e, ok := err.(coder); ok {
		return Errorf(e.Code(), err.Error())
	} else if e, ok := err.(*types.Error); ok {
		return Errorf(e.Code, e.Message)
	} else if err == types.ErrNotExist {
		return Errorf(http.StatusNotFound, err.Error())
	} else {
		return Errorf(http.StatusInternalServerError, err.Error())
	}
}
