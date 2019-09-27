package wine

import (
	"net/http"

	"github.com/gopub/wine/internal/request"
	"github.com/gopub/wine/mime"

	"github.com/gopub/gox"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	request     *http.Request
	params      gox.M
	body        []byte
	contentType string
}

func (r *Request) Request() *http.Request {
	return r.request
}

func (r *Request) Params() gox.M {
	return r.params
}

func (r *Request) Body() []byte {
	return r.body
}

func (r *Request) ContentType() string {
	return r.contentType
}

func NewRequest(r *http.Request, parser ParamsParser) (*Request, error) {
	if parser == nil {
		parser = request.NewParamsParser(nil, 8*gox.MB)
	}

	params, body, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Request{
		request:     r,
		params:      params,
		body:        body,
		contentType: mime.GetContentType(r.Header),
	}, nil
}

type ParamsParser interface {
	Parse(req *http.Request) (gox.M, []byte, error)
}
