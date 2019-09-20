package api

import (
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
	"net/http"
)

type errorInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type responseInfo struct {
	Error *errorInfo  `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func Data(data interface{}) wine.Responsible {
	header := make(http.Header)
	header.Set(wine.ContentType, mime.JSON)
	val := &responseInfo{
		Data: data,
	}
	return wine.NewResponse(http.StatusOK, header, val)
}

func StatusData(status int, data interface{}) wine.Responsible {
	header := make(http.Header)
	header.Set(wine.ContentType, mime.JSON)
	val := &responseInfo{
		Data: data,
	}
	return wine.NewResponse(status, header, val)
}

func Error(code int, message string) wine.Responsible {
	header := make(http.Header)
	header.Set(wine.ContentType, mime.JSON)
	val := &responseInfo{
		Error: &errorInfo{
			Code:    code,
			Message: message,
		},
	}
	status := code
	for status >= 1000 {
		status /= 10
	}
	return wine.NewResponse(status, header, val)
}
