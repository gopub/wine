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
	"github.com/gopub/wine/httpvalue"
)

type RequestParams struct {
	CookieParams types.M
	HeaderParams types.M
	QueryParams  types.M
	PathParams   types.M
	BodyParams   types.M
}

func (p *RequestParams) Combine() types.M {
	params := types.M{}
	params.AddMap(p.CookieParams)
	params.AddMap(p.HeaderParams)
	params.AddMap(p.PathParams)
	params.AddMap(p.QueryParams)
	params.AddMap(p.BodyParams)
	return params
}

func ReadRequest(req *http.Request, maxMemory types.ByteUnit) (*RequestParams, []byte, error) {
	params := &RequestParams{
		CookieParams: ReadCookies(req.Cookies()),
		HeaderParams: ReadHeader(req.Header),
		QueryParams:  ReadValues(req.URL.Query()),
	}
	bp, body, err := ReadBody(req, maxMemory)
	if err != nil {
		return params, body, fmt.Errorf("read request body: %w", err)
	}
	params.BodyParams = bp
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
	for k, va := range values {
		isArray := strings.HasSuffix(k, "[]")
		if isArray {
			k = k[0 : len(k)-2]
			if k == "" {
				continue
			}

			if len(va) == 1 {
				va = strings.Split(va[0], ",")
			}
		}

		if len(va) == 0 {
			continue
		}

		k = strings.ToLower(k)
		if isArray || len(va) > 1 {
			// value is an array or expected to be an array
			m[k] = va
		} else {
			m[k] = va[0]
		}
	}

	if jsonStr := m.String("json"); jsonStr != "" {
		var j types.M
		err := json.Unmarshal([]byte(jsonStr), &j)
		if err == nil {
			j.AddMap(m)
			m = j
		}
	}
	return m
}

func ReadBody(req *http.Request, maxMemory types.ByteUnit) (types.M, []byte, error) {
	typ := httpvalue.GetContentType(req.Header)
	params := types.M{}
	switch typ {
	case httpvalue.HTML, httpvalue.Plain:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return params, nil, fmt.Errorf("read html or plain body: %w", err)
		}
		return params, body, nil
	case httpvalue.JSON:
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
	case httpvalue.FormURLEncoded:
		// TODO: will crash
		//body, err := req.GetBody()
		//if err != nil {
		//	return params, nil, fmt.Errorf("get body: %w", err)
		//}
		//bodyData, err := ioutil.Read(body)
		//body.Close()
		//if err != nil {
		//	return params, nil, fmt.Errorf("read form body: %w", err)
		//}
		if err := req.ParseForm(); err != nil {
			return params, nil, fmt.Errorf("parse form: %w", err)
		}
		return ReadValues(req.Form), nil, nil
	case httpvalue.FormData:
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
