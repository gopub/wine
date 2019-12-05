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

// Request returns original http request
func (r *Request) Request() *http.Request {
	return r.request
}

// Params returns request parameters
func (r *Request) Params() gox.M {
	return r.params
}

// Body returns request body
func (r *Request) Body() []byte {
	return r.body
}

// ContentType returns request's content type
func (r *Request) ContentType() string {
	return r.contentType
}

// SessionID returns request's session id
func (r *Request) SessionID() string {
	return r.params.String(SessionName)
}

// Authorization returns request's Authorization in header
func (r *Request) Authorization() string {
	return r.request.Header.Get("Authorization")
}

// Bearer returns bearer token in header
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

func newRequest(r *http.Request, parser ParamsParser) (*Request, error) {
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

// ParamsParser interface is used by Wine server to parse parameters from http request
type ParamsParser interface {
	Parse(req *http.Request) (gox.M, []byte, error)
}
