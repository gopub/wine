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
	HTTPRequest *http.Request
	Parameters  gox.M
}

type RequestParser interface {
	ParseHTTPRequest(req *http.Request, maxMemory int64) (parameters gox.M, err error)
}

type DefaultRequestParser struct {
	headerFields map[string]bool
}

func NewDefaultRequestParser() *DefaultRequestParser {
	return &DefaultRequestParser{
		headerFields: map[string]bool{
			"sid":       true,
			"device_id": true,
		},
	}
}

func (p *DefaultRequestParser) ParseHTTPRequest(req *http.Request, maxMemory int64) (gox.M, error) {
	params := gox.M{}
	params.AddMap(parseCookieParams(req))
	params.AddMap(parseHeaderParams(req))
	params.AddMap(parseURLValues(req.URL.Query()))
	bp, err := parseBodyParams(req, maxMemory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse body")
	}
	params.AddMap(bp)
	return params, nil
}

func parseCookieParams(req *http.Request) gox.M {
	params := gox.M{}
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}
	return params
}

func parseHeaderParams(req *http.Request) gox.M {
	params := gox.M{}
	for k, v := range req.Header {
		if strings.HasPrefix(k, "x-") || strings.HasPrefix(k, "X-") || p.headerFields[k] {
			params[strings.ToLower(k[2:])] = v
		}
	}
	return params
}

func parseURLValues(values url.Values) gox.M {
	m := gox.M{}
	for k, v := range values {
		i := strings.Index(k, "[]")
		if i >= 0 && i == len(k)-2 {
			k = k[0 : len(k)-2]
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

func parseContentType(req *http.Request) string {
	t := req.Header.Get(ContentType)
	for i, ch := range t {
		if ch == ' ' || ch == ';' {
			t = t[:i]
			break
		}
	}
	return t
}

func parseBodyParams(req *http.Request, maxMemory int64) (gox.M, error) {
	typ := parseContentType(req)
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
		return parseURLValues(req.Form), nil
	case mime.FormData:
		err := req.ParseMultipartForm(maxMemory)
		if err != nil {
			return nil, errors.Wrap(err, "cannot parse multipart form")
		}

		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			return parseURLValues(req.MultipartForm.Value), nil
		}
		return params, nil
	default:
		if len(typ) != 0 {
			logger.Warnf("Ignored content type=%s", typ)
		}
		return params, nil
	}
}
