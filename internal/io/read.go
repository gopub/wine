package io

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopub/types"
	"github.com/gopub/wine/mime"
)

func ReadRequest(req *http.Request, maxMemory types.ByteUnit) (types.M, []byte, error) {
	params := types.M{}
	params.AddMap(ReadCookies(req.Cookies()))
	params.AddMap(ReadHeader(req.Header))
	params.AddMap(ReadValues(req.URL.Query()))
	bp, body, err := ReadBody(req, maxMemory)
	if err != nil {
		return params, body, fmt.Errorf("read request body: %w", err)
	}
	params.AddMap(bp)
	return params, body, nil
}

func ReadCookies(cookies []*http.Cookie) types.M {
	params := types.M{}
	for _, c := range cookies {
		params[c.Name] = c.Value
	}
	return params
}

func ReadHeader(h http.Header) types.M {
	params := types.M{}
	for k, v := range h {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-") {
			k = k[2:]
			k = strings.Replace(k, "-", "_", -1)
			params[k] = v
		}
	}
	return params
}

func ReadValues(values url.Values) types.M {
	m := types.M{}
	for k, v := range values {
		i := strings.Index(k, "[]")
		if i >= 0 && i == len(k)-2 {
			k = k[0 : len(k)-2]
			if len(v) == 1 {
				v = strings.Split(v[0], ",")
			}
		}
		k = strings.ToLower(k)
		if len(v) > 1 || i >= 0 {
			m[k] = v
		} else if len(v) == 1 {
			m[k] = v[0]
		}
	}

	return m
}

func ReadBody(req *http.Request, maxMemory types.ByteUnit) (types.M, []byte, error) {
	typ := mime.GetContentType(req.Header)
	params := types.M{}
	switch typ {
	case mime.HTML, mime.Plain:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return params, nil, fmt.Errorf("read html or plain body: %w", err)
		}
		return params, body, nil
	case mime.JSON:
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return params, nil, fmt.Errorf("read json body: %w", err)
		}
		if len(body) == 0 {
			return params, nil, nil
		}
		decoder := json.NewDecoder(bytes.NewBuffer(body))
		decoder.UseNumber()
		err = decoder.Decode(&params)
		if err != nil {
			var obj interface{}
			err = json.Unmarshal(body, &obj)
			if err != nil {
				return params, body, fmt.Errorf("unmarshal json %s: %w", string(body), err)
			}
		}
		return params, body, nil
	case mime.FormURLEncoded:
		// TODO: will crash
		//body, err := req.GetBody()
		//if err != nil {
		//	return params, nil, fmt.Errorf("get body: %w", err)
		//}
		//bodyData, err := ioutil.ReadAll(body)
		//body.Close()
		//if err != nil {
		//	return params, nil, fmt.Errorf("read form body: %w", err)
		//}
		if err := req.ParseForm(); err != nil {
			return params, nil, fmt.Errorf("parse form: %w", err)
		}
		return ReadValues(req.Form), nil, nil
	case mime.FormData:
		err := req.ParseMultipartForm(int64(maxMemory))
		if err != nil {
			return nil, nil, fmt.Errorf("parse multipart form: %w", err)
		}

		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			return ReadValues(req.MultipartForm.Value), nil, nil
		}
		return params, nil, nil
	default:
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return params, nil, fmt.Errorf("read json body: %w", err)
		}
		return params, body, nil
	}
}
