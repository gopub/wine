package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/log"
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

type Result struct {
	Error *gox.Error  `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func Data(data interface{}) wine.Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.JSON)
	val := &Result{
		Data: data,
	}
	return wine.NewResponse(http.StatusOK, header, val)
}

func StatusData(status int, data interface{}) wine.Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.JSON)
	val := &Result{
		Data: data,
	}
	return wine.NewResponse(status, header, val)
}

func ErrorMessage(code int, message string) wine.Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.JSON)
	val := &Result{
		Error: gox.NewError(code, message),
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
func ParseResult(resp *http.Response, dataModel interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return errors.Wrap(err, "read response body failed")
	}

	ct := mime.GetContentType(resp.Header)
	switch ct {
	case mime.JSON:
		res := new(Result)
		// Use dataModel (usually is a pointer to a struct) to hold decoded data
		res.Data = dataModel
		if err = json.Unmarshal(body, res); err != nil {
			log.Errorf("Unmarshal response body failed: %s %v", string(body), err)
			return gox.NewError(StatusInvalidResponse, err.Error())
		}

		if res.Error != nil {
			return res.Error
		}
	default:
		break
	}
	if resp.StatusCode >= http.StatusBadRequest {
		ge := gox.NewError(resp.StatusCode, "")
		if len(body) < 256 {
			ge.Message = string(body)
		}
		return ge
	}
	return nil
}
