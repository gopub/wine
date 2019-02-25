package wine

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gopub/gox"
	"github.com/gopub/wine/mime"
	"io/ioutil"
	"net/http"
	"strings"
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
	for _, cookie := range req.Cookies() {
		params[cookie.Name] = cookie.Value
	}

	for k, v := range req.Header {
		if strings.HasPrefix(k, "x-") || strings.HasPrefix(k, "X-") || p.headerFields[k] {
			params[strings.ToLower(k[2:])] = v
		}
	}

	params.AddMap(convertToM(req.URL.Query()))

	contentType := req.Header.Get(ContentType)
	for i, ch := range contentType {
		if ch == ' ' || ch == ';' {
			contentType = contentType[:i]
			break
		}
	}

	switch contentType {
	case mime.HTML, mime.Plain:
		break
	case mime.JSON:
		d, e := ioutil.ReadAll(req.Body)
		if e != nil {
			logger.Error(e)
			break
		}

		if len(d) > 0 {
			var m gox.M
			e = jsonUnmarshal(d, &m)
			if e != nil {
				break
			}
			params.AddMap(m)
		}
	case mime.FormURLEncoded:
		//startAt := time.Now()
		err := req.ParseForm()
		//logger.Debug("Cost:", time.Since(startAt))
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		params.AddMap(convertToM(req.Form))
	case mime.FormData:
		//startAt := time.Now()
		// ParseMultipartForm is very slow sometimes. Why???
		err := req.ParseMultipartForm(maxMemory)
		//logger.Debug("Cost:", time.Since(startAt))
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			params.AddMap(convertToM(req.MultipartForm.Value))
		}
	default:
		logger.Infof("unprocessed content type=%s", contentType)
	}

	return params, nil
}

func convertToM(values map[string][]string) gox.M {
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

func jsonUnmarshal(data []byte, pJSONObj interface{}) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(pJSONObj)
	if err != nil {
		logger.Error(err)
	}
	return err
}
