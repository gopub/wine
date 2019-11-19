package wine

import (
	"net/http"
	"strings"

	"github.com/gopub/gox"
	"github.com/gopub/wine/internal/request"
	"github.com/gopub/wine/mime"
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

func (r *Request) SessionID() string {
	return r.params.String(SessionName)
}

func (r *Request) Authorization() string {
	return r.request.Header.Get("Authorization")
}

func (r *Request) Bearer() string {
	s := r.Authorization()
	strs := strings.Split(s, " ")
	if len(strs) != 2 {
		return ""
	}
	if strs[0] == "Bearer" {
		return strs[1]
	}
	return ""
}

func NewRequest(r *http.Request, parser ParamsParser) (*Request, error) {
	if parser == nil {
		parser = request.NewParamsParser(8 * gox.MB)
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
