package wine

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gopub/log"
	"github.com/gopub/wine/mime"
)

// Responder interface is used by Wine server to write response to the client
type Responder interface {
	// Respond will be called to write status/body to http response writer
	Respond(ctx context.Context, w http.ResponseWriter)
}

var (
	_ Handler   = ResponderFunc(nil)
	_ Responder = ResponderFunc(nil)
)

// ResponderFunc is a func that implements interface Responder
type ResponderFunc func(ctx context.Context, w http.ResponseWriter)

func (f ResponderFunc) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}

func (f ResponderFunc) HandleRequest(_ context.Context, _ *Request) Responder {
	return f
}

var (
	_ Handler   = (*Response)(nil)
	_ Responder = (*Response)(nil)
)

// Response holds all the http response information
// Value and headers except the status code can be modified before sent to the client
type Response struct {
	status int
	header http.Header
	value  interface{}
}

// Respond writes header and body to response writer w
func (r *Response) Respond(_ context.Context, w http.ResponseWriter) {
	body, ok := r.value.([]byte)
	if !ok {
		body = r.getBytes()
	}

	for k, v := range r.header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.status)
	if _, err := w.Write(body); err != nil {
		log.Error(err)
	}
}

func (r *Response) HandleRequest(ctx context.Context, req *Request) Responder {
	return r
}

func (r *Response) getBytes() []byte {
	if body, ok := r.value.([]byte); ok {
		return body
	}

	contentType := r.header.Get(mime.ContentType)

	switch {
	case strings.Contains(contentType, mime.JSON):
		if r.value != nil {
			body, err := json.Marshal(r.value)
			if err != nil {
				logger.Error(err)
			}
			return body
		}
	case strings.Contains(contentType, mime.Plain):
		fallthrough
	case strings.Contains(contentType, mime.HTML):
		fallthrough
	case strings.Contains(contentType, mime.XML):
		fallthrough
	case strings.Contains(contentType, mime.XML2):
		if s, ok := r.value.(string); ok {
			return []byte(s)
		}
	default:
		log.Warn("Unsupported Content-Type:", contentType)
	}

	return nil
}

func (r *Response) Status() int {
	return r.status
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) Value() interface{} {
	return r.value
}

func (r *Response) SetValue(v interface{}) {
	r.value = v
}

var OK = Status(http.StatusOK)

// Status returns a response only with a status code
func Status(status int) *Response {
	return Text(status, http.StatusText(status))
}

// Redirect sends a redirect response
func Redirect(location string, permanent bool) *Response {
	header := make(http.Header)
	header.Set("Location", location)
	header.Set(mime.ContentType, mime.Plain)
	var status int
	if permanent {
		status = http.StatusMovedPermanently
	} else {
		status = http.StatusFound
	}

	return &Response{
		status: status,
		header: header,
	}
}

// Text sends a text response
func Text(status int, text string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.PlainUTF8)
	return &Response{
		status: status,
		header: header,
		value:  text,
	}
}

// JSON creates a application/json response
func JSON(status int, value interface{}) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.JsonUTF8)
	return &Response{
		status: status,
		header: header,
		value:  value,
	}
}

// Handle handles request with h
func Handle(req *http.Request, h http.Handler) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		h.ServeHTTP(w, req)
	})
}
