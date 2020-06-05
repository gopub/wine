package io

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/wine/mime"
)

func DecodeResponse(resp *http.Response, result interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("read resp body: %v", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return errors.Format(resp.StatusCode, string(body))
	}
	if result == nil {
		return nil
	}
	ct := mime.GetContentType(resp.Header)
	switch {
	case strings.Contains(ct, mime.JSON):
		return json.Unmarshal(body, result)
	case strings.Contains(ct, mime.Protobuf):
		m, ok := result.(proto.Message)
		if !ok {
			return fmt.Errorf("expected proto.Message instead of %T", result)
		}
		return proto.Unmarshal(body, m)
	case strings.Contains(ct, mime.Plain):
		if len(body) == 0 {
			return errors.New("no data")
		}
		return conv.SetBytes(result, body)
	default:
		return errors.New("invalid result")
	}
}
