package api

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gopub/log"
	"github.com/gopub/types"
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
	Error *types.Error `json:"error,omitempty"`
	Data  interface{}  `json:"data"`
}

var OK = Data(nil)

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

// ParseResult parse response at client side
func ParseResult(resp *http.Response, dataModel interface{}, useResultModel bool) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Read response body: %v", err)
		return types.NewError(StatusTransportFailed, fmt.Sprintf("read response body: %v", err))
	}

	ct := mime.GetContentType(resp.Header)
	switch ct {
	case mime.JSON:
		target := dataModel
		if useResultModel {
			res := new(Result)
			// Use dataModel (usually is a pointer to a struct) to hold decoded data
			res.Data = dataModel
			target = res
		}

		if target != nil {
			if err = json.Unmarshal(body, target); err != nil {
				const logLimit = 512
				bodyStr := string(body)
				if len(bodyStr) > logLimit {
					bodyStr = bodyStr[:logLimit] + "..."
				}
				log.Errorf("Unmarshal response body: %s %v", bodyStr, err)
				if resp.StatusCode >= http.StatusBadRequest {
					return types.NewError(resp.StatusCode, bodyStr)
				}
				return types.NewError(StatusInvalidResponse, fmt.Sprintf("unmarshal json body: %v", err))
			}
		}

		if res, ok := target.(*Result); ok && res.Error != nil {
			return res.Error
		}
		return nil
	case mime.Plain:
		if resp.StatusCode >= http.StatusBadRequest {
			return types.NewError(resp.StatusCode, string(body))
		}
		if dataModel == nil {
			return nil
		}
		if len(body) == 0 {
			return types.NewError(StatusInvalidResponse, "no data")
		}
		return assign(dataModel, body)
	default:
		if resp.StatusCode >= http.StatusBadRequest {
			return types.NewError(resp.StatusCode, string(body))
		}

		if dataModel != nil {
			return types.NewError(StatusInvalidResponse, "no data")
		}
		return nil
	}
}

func assign(dataModel interface{}, body []byte) error {
	v := reflect.ValueOf(dataModel)
	if v.Kind() != reflect.Ptr {
		log.Panicf("Argument dataModel %T is not pointer", dataModel)
	}

	elem := v.Elem()
	if !elem.CanSet() {
		log.Panicf("Argument dataModel %T cannot be set", dataModel)
	}

	if tu, ok := v.Interface().(encoding.TextUnmarshaler); ok {
		err := tu.UnmarshalText(body)
		if err != nil {
			return types.NewError(StatusInvalidResponse, fmt.Sprintf("unmarshal text: %v", err))
		}
		return nil
	}

	switch elem.Kind() {
	case reflect.String:
		elem.SetString(string(body))
	case reflect.Int64,
		reflect.Int32,
		reflect.Int,
		reflect.Int16,
		reflect.Int8:
		i, err := strconv.ParseInt(string(body), 10, 64)
		if err != nil {
			return types.NewError(StatusInvalidResponse, fmt.Sprintf("parse int: %v", err))
		}
		elem.SetInt(i)
	case reflect.Uint64,
		reflect.Uint32,
		reflect.Uint,
		reflect.Uint16,
		reflect.Uint8:
		i, err := strconv.ParseUint(string(body), 10, 64)
		if err != nil {
			return types.NewError(StatusInvalidResponse, fmt.Sprintf("parse uint: %v", err))
		}
		elem.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := strconv.ParseFloat(string(body), 64)
		if err != nil {
			return types.NewError(StatusInvalidResponse, fmt.Sprintf("parse float: %v", err))
		}
		elem.SetFloat(i)
	case reflect.Bool:
		i, err := strconv.ParseBool(string(body))
		if err != nil {
			return types.NewError(StatusInvalidResponse, fmt.Sprintf("parse bool: %v", err))
		}
		elem.SetBool(i)
	default:
		return fmt.Errorf("cannot assign to dataModel %T", dataModel)
	}
	return nil
}
