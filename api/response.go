package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

type coder interface {
	Code() int
}

type messageCoder interface {
	coder
	Message() string
}

type Result struct {
	Error *gox.Error  `json:"error,omitempty"`
	Data  interface{} `json:"data"`
}

func Data(data interface{}) wine.Responsible {
	val := &Result{
		Data: data,
	}
	return wine.JSON(http.StatusOK, val)
}

func StatusData(status int, data interface{}) wine.Responsible {
	val := &Result{
		Data: data,
	}
	return wine.JSON(status, val)
}

func ErrorMessage(code int, message string) wine.Responsible {
	val := &Result{
		Error: gox.NewError(code, message),
	}
	status := code
	for status >= 1000 {
		status /= 10
	}
	return wine.JSON(status, val)
}

func Error(err error) wine.Responsible {
	err = gox.Cause(err)
	if e, ok := err.(messageCoder); ok {
		return ErrorMessage(e.Code(), e.Message())
	} else if e, ok := err.(coder); ok {
		return ErrorMessage(e.Code(), err.Error())
	} else if e, ok := err.(*gox.Error); ok {
		return ErrorMessage(e.Code, e.Message)
	} else if err == gox.ErrNotExist {
		return ErrorMessage(http.StatusNotFound, err.Error())
	} else {
		return ErrorMessage(http.StatusInternalServerError, err.Error())
	}
}

// ParseResponse parse response at client side
func ParseResult(resp *http.Response, dataModel interface{}, useResultModel bool) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Read response body: %v", err)
		return gox.NewError(StatusTransportFailed, fmt.Sprintf("read response body: %v", err))
	}

	ct := mime.GetContentType(resp.Header)
	switch ct {
	case mime.JSON:
		if useResultModel {
			res := new(Result)
			// Use dataModel (usually is a pointer to a struct) to hold decoded data
			res.Data = dataModel
			if err = json.Unmarshal(body, res); err != nil {
				log.Errorf("Unmarshal response body: %s %v", string(body), err)
				if resp.StatusCode >= http.StatusBadRequest {
					return gox.NewError(resp.StatusCode, string(body))
				}
				return gox.NewError(StatusInvalidResponse, fmt.Sprintf("unmarshal json response: %v", err))
			}

			if res.Error != nil {
				return res.Error
			}
		} else {
			if dataModel != nil {
				if err = json.Unmarshal(body, dataModel); err != nil {
					log.Errorf("Unmarshal response body: %s %v", string(body), err)
					if resp.StatusCode >= http.StatusBadRequest {
						return gox.NewError(resp.StatusCode, string(body))
					}
					return gox.NewError(StatusInvalidResponse, fmt.Sprintf("unmarshal json response: %v", err))
				}
			}
		}
		return nil
	default:
		if resp.StatusCode >= http.StatusBadRequest {
			return gox.NewError(resp.StatusCode, string(body))
		}

		if dataModel != nil {
			return gox.NewError(StatusInvalidResponse, "no data")
		}
		return nil
	}
}
