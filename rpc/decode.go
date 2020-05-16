package rpc

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/conv"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/mime"
)

const bodyLogMaxLen = 512

type Decoder interface {
	Decode(resp *http.Response, data interface{}) error
}

type jsonResult struct {
	Error *types.Error `json:"error,omitempty"`
	Data  interface{}  `json:"data"`
}

type StdDecoder struct {
}

func (r *StdDecoder) Decode(resp *http.Response, data interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("read resp body: %v", err)
	}
	status := resp.StatusCode
	ct := mime.GetContentType(resp.Header)
	switch {
	case strings.Contains(ct, mime.JSON):
		return r.decodeJSON(status, body, data)
	case strings.Contains(ct, mime.Protobuf):
		return r.decodeProtobuf(status, body, data)
	case strings.Contains(ct, mime.Plain):
		return r.decodePlainText(status, body, data)
	default:
		if resp.StatusCode >= http.StatusBadRequest {
			return types.NewError(status, string(body))
		}

		if data != nil {
			return errors.New("no data")
		}
		return nil
	}
}

func (r *StdDecoder) decodeJSON(status int, body []byte, data interface{}) error {
	res := &jsonResult{Data: data}
	err := json.Unmarshal(body, res)
	if err != nil {
		s := trimCount(string(body), bodyLogMaxLen)
		if status >= http.StatusBadRequest {
			return types.NewError(status, s)
		}
		return fmt.Errorf("unmarshal: %v", err)
	}
	return res.Error
}

func (r *StdDecoder) decodeProtobuf(status int, body []byte, data interface{}) error {
	if data == nil {
		if status >= http.StatusBadRequest {
			s := trimCount(string(body), bodyLogMaxLen)
			if status >= http.StatusBadRequest {
				return types.NewError(status, s)
			}
		}
		return nil
	}

	m, ok := data.(proto.Message)
	if !ok {
		return fmt.Errorf("expected proto.Message instead of %T", data)
	}

	err := proto.Unmarshal(body, m)
	if err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}
	return nil
}

func (r *StdDecoder) decodePlainText(status int, body []byte, data interface{}) error {
	if status >= http.StatusBadRequest {
		return types.NewError(status, string(body))
	}
	if data == nil {
		return nil
	}
	if len(body) == 0 {
		return errors.New("no data")
	}
	return assign(data, body)
}

func trimCount(s string, n int) string {
	if len(s) > n {
		return s[:bodyLogMaxLen] + "..."
	}
	return s
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
			return fmt.Errorf("unmarshal text: %w", err)
		}
		return nil
	}

	if bu, ok := v.Interface().(encoding.BinaryUnmarshaler); ok {
		err := bu.UnmarshalBinary(body)
		if err != nil {
			return fmt.Errorf("unmarshal binary: %w", err)
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
		i, err := conv.ToInt64(body)
		if err != nil {
			return fmt.Errorf("parse int: %v", err)
		}
		elem.SetInt(i)
	case reflect.Uint64,
		reflect.Uint32,
		reflect.Uint,
		reflect.Uint16,
		reflect.Uint8:
		i, err := conv.ToUint64(body)
		if err != nil {
			return fmt.Errorf("parse uint: %w", err)
		}
		elem.SetUint(i)
	case reflect.Float32, reflect.Float64:
		i, err := conv.ToFloat64(body)
		if err != nil {
			return fmt.Errorf("parse float: %w", err)
		}
		elem.SetFloat(i)
	case reflect.Bool:
		i, err := conv.ToBool(body)
		if err != nil {
			return fmt.Errorf("parse bool: %w", err)
		}
		elem.SetBool(i)
	default:
		return fmt.Errorf("cannot assign to dataModel %T", dataModel)
	}
	return nil
}
