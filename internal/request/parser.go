package request

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopub/gox"
	"github.com/gopub/wine/mime"
	"github.com/pkg/errors"
)

type ParamsParser struct {
	headerParamNames *gox.StringSet
	maxMemory        gox.ByteUnit
}

func NewParamsParser(headerParamNames []string, maxMemory gox.ByteUnit) *ParamsParser {
	p := &ParamsParser{
		headerParamNames: gox.NewStringSet(1),
		maxMemory:        maxMemory,
	}
	for _, n := range headerParamNames {
		p.headerParamNames.Add(n)
	}
	if p.maxMemory < gox.MB {
		p.maxMemory = gox.MB
	}
	return p
}

func (p *ParamsParser) Parse(req *http.Request) (gox.M, []byte, error) {
	params := gox.M{}
	params.AddMap(p.parseCookie(req))
	params.AddMap(p.parseHeader(req))
	params.AddMap(p.parseURLValues(req.URL.Query()))
	bp, body, err := p.parseBody(req)
	if err != nil {
		return params, body, errors.Wrap(err, "parse body failed")
	}
	params.AddMap(bp)
	return params, body, nil
}

func (p *ParamsParser) parseCookie(req *http.Request) gox.M {
	params := gox.M{}
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}
	return params
}

func (p *ParamsParser) parseHeader(req *http.Request) gox.M {
	params := gox.M{}
	for k, v := range req.Header {
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-") {
			k = k[2:]
		}

		if p.headerParamNames.Contains(k) {
			params[k] = v
		}

		k = strings.Replace(k, "-", "_", -1)
		if p.headerParamNames.Contains(k) {
			params[k] = v
		}
	}
	return params
}

func (p *ParamsParser) parseURLValues(values url.Values) gox.M {
	m := gox.M{}
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

func (p *ParamsParser) parseBody(req *http.Request) (gox.M, []byte, error) {
	typ := mime.GetContentType(req.Header)
	params := gox.M{}
	switch typ {
	case mime.HTML, mime.Plain:
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return params, nil, errors.Wrap(err, "read html or plain body failed")
		}
		return params, body, nil
	case mime.JSON:
		body, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return params, nil, errors.Wrap(err, "read json body failed")
		}
		if len(body) == 0 {
			return params, nil, nil
		}
		decoder := json.NewDecoder(bytes.NewBuffer(body))
		decoder.UseNumber()
		err = decoder.Decode(&params)
		return params, body, errors.Wrapf(err, "decode json failed: %s", string(body))
	case mime.FormURLEncoded:
		body, err := req.GetBody()
		if err != nil {
			return params, nil, errors.Wrap(err, "get body failed")
		}
		bodyData, err := ioutil.ReadAll(body)
		body.Close()
		if err != nil {
			return params, nil, errors.Wrap(err, "read form body failed")
		}
		if err = req.ParseForm(); err != nil {
			return params, bodyData, errors.Wrap(err, "parse form failed")
		}
		return p.parseURLValues(req.Form), bodyData, nil
	case mime.FormData:
		err := req.ParseMultipartForm(int64(p.maxMemory))
		if err != nil {
			return nil, nil, errors.Wrap(err, "parse multipart form failed")
		}

		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			return p.parseURLValues(req.MultipartForm.Value), nil, nil
		}
		return params, nil, nil
	default:
		return params, nil, nil
	}
}
