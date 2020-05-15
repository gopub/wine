package respond

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gopub/log"
	"github.com/gopub/wine/mime"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

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

type Func func(ctx context.Context, w http.ResponseWriter)

func (f Func) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}
