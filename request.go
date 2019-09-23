package wine

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine/mime"
	"github.com/pkg/errors"
)

const (
	ContentType = "Content-Type"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	request *http.Request
	params  gox.M
}

func (r *Request) Request() *http.Request {
	return r.request
}

func (r *Request) Params() gox.M {
	return r.params
}

func NewRequest(r *http.Request, parser ParamsParser) (*Request, error) {
	if parser == nil {
		parser = NewDefaultParamsParser(nil, 8*gox.MB)
	}

	params, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Request{
		request: r,
		params:  params,
	}, nil
}

type ParamsParser interface {
	Parse(req *http.Request) (gox.M, error)
}

type DefaultParamsParser struct {
	headerParamNames *gox.StringSet
	maxMemory        gox.ByteUnit
}

func NewDefaultParamsParser(headerParamNames []string, maxMemory gox.ByteUnit) *DefaultParamsParser {
	p := &DefaultParamsParser{
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

func (p *DefaultParamsParser) Parse(req *http.Request) (gox.M, error) {
	params := gox.M{}
	params.AddMap(p.parseCookieParams(req))
	params.AddMap(p.parseHeaderParams(req))
	params.AddMap(p.parseURLValues(req.URL.Query()))
	bp, err := p.parseBodyParams(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse body")
	}
	params.AddMap(bp)
	return params, nil
}

func (p *DefaultParamsParser) parseCookieParams(req *http.Request) gox.M {
	params := gox.M{}
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}
	return params
}

func (p *DefaultParamsParser) parseHeaderParams(req *http.Request) gox.M {
	params := gox.M{}
	for k, v := range req.Header {
		if strings.HasPrefix(k, "x-") || strings.HasPrefix(k, "X-") || p.headerParamNames.Contains(k) {
			params[strings.ToLower(k[2:])] = v
		}
	}
	return params
}

func (p *DefaultParamsParser) parseURLValues(values url.Values) gox.M {
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

func (p *DefaultParamsParser) parseContentType(req *http.Request) string {
	t := req.Header.Get(ContentType)
	for i, ch := range t {
		if ch == ' ' || ch == ';' {
			t = t[:i]
			break
		}
	}
	return t
}

func (p *DefaultParamsParser) parseBodyParams(req *http.Request) (gox.M, error) {
	typ := p.parseContentType(req)
	params := gox.M{}
	switch typ {
	case mime.HTML, mime.Plain:
		return params, nil
	case mime.JSON:
		decoder := json.NewDecoder(req.Body)
		decoder.UseNumber()
		defer func() {
			if err := req.Body.Close(); err != nil {
				log.Errorf("cannot close request body: %v", err)
			}
		}()

		err := decoder.Decode(&params)
		if err != nil {
			return nil, errors.Wrap(err, "cannot decode")
		}
		return params, nil
	case mime.FormURLEncoded:
		err := req.ParseForm()
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse form")
		}
		return p.parseURLValues(req.Form), nil
	case mime.FormData:
		err := req.ParseMultipartForm(int64(p.maxMemory))
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse multipart form")
		}

		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			return p.parseURLValues(req.MultipartForm.Value), nil
		}
		return params, nil
	default:
		if len(typ) != 0 {
			logger.Warnf("Ignored content type=%s", typ)
		}
		return params, nil
	}
}
