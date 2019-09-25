package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gopub/log"

	"github.com/gopub/gox"

	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
	"github.com/pkg/errors"
)

type coder interface {
	Code() int
}

type coderMessager interface {
	coder
	Message() string
}

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

func ErrorMessage(code int, message string) wine.Responsible {
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

func Error(err error) wine.Responsible {
	err = errors.Cause(err)
	if e, ok := err.(coderMessager); ok {
		return ErrorMessage(e.Code(), e.Message())
	} else if e, ok := err.(coder); ok {
		return ErrorMessage(e.Code(), err.Error())
	} else if e, ok := err.(*gox.Error); ok {
		return ErrorMessage(e.Code, e.Message)
	} else {
		return ErrorMessage(http.StatusInternalServerError, err.Error())
	}
}

// ParseResponse parse response at client side
func ParseResponse(resp *http.Response, result interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return errors.Wrap(err, "read response body failed")
	}

	ct := wine.GetContentType(resp.Header)
	switch ct {
	case mime.HTML, mime.Plain, mime.PlainContentType:
		if resp.StatusCode < http.StatusMultipleChoices {
			return nil
		}
		return gox.NewError(resp.StatusCode, string(body))
	case mime.JSON:
		info := new(responseInfo)
		// Use result (usually is a pointer to a struct) to hold decoded data
		info.Data = result
		if err = json.Unmarshal(body, info); err != nil {
			log.Errorf("Unmarshal response body failed: %s %v", string(body), err)
			return gox.NewError(StatusInvalidResponse, err.Error())
		}

		if info.Error != nil {
			return gox.NewError(info.Error.Code, info.Error.Message)
		}
		return nil
	default:
		break
	}
	return nil
}
