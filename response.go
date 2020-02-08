package wine

import (
	"context"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine/mime"
)

const (
	charsetSuffix = "; charset=utf-8"
	ctPlain       = mime.Plain + charsetSuffix
	ctHTML        = mime.HTML + charsetSuffix
	ctJSON        = mime.JSON + charsetSuffix
	ctXML         = mime.XML + charsetSuffix
)

// Responsible interface is used by Wine server to write response to the client
type Responsible interface {
	Respond(ctx context.Context, w http.ResponseWriter)
}

// ResponsibleFunc is a func that implements interface  Responsible
type ResponsibleFunc func(ctx context.Context, w http.ResponseWriter)

// Respond will be called to write status/body to http response writer
func (f ResponsibleFunc) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}

// Response defines the http response interface
type Response interface {
	Responsible
	Status() int
	Header() http.Header
	Value() interface{}
	SetValue(v interface{})
}

type responseImpl struct {
	status int
	header http.Header
	value  interface{}
}

// NewResponse creates a new response
func NewResponse(status int, header http.Header, value interface{}) Response {
	return &responseImpl{
		status: status,
		header: header,
		value:  value,
	}
}

// Respond writes header and body to response writer w
func (r *responseImpl) Respond(ctx context.Context, w http.ResponseWriter) {
	body, ok := r.value.([]byte)
	if !ok {
		body = r.getBytes()
	}

	for k, v := range r.header {
		w.Header()[k] = v
	}
	w.WriteHeader(r.status)
	if err := gox.WriteAll(w, body); err != nil {
		log.Error(err)
	}
}

func (r *responseImpl) getBytes() []byte {
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

func (r *responseImpl) Status() int {
	return r.status
}

func (r *responseImpl) Header() http.Header {
	return r.header
}

func (r *responseImpl) Value() interface{} {
	return r.value
}

func (r *responseImpl) SetValue(v interface{}) {
	r.value = v
}

// Status returns a response only with a status code
func Status(status int) Responsible {
	return Text(status, http.StatusText(status))
}

// Redirect sends a redirect response
func Redirect(location string, permanent bool) Responsible {
	header := make(http.Header)
	header.Set("Location", location)
	header.Set(mime.ContentType, mime.Plain)
	var status int
	if permanent {
		status = http.StatusMovedPermanently
	} else {
		status = http.StatusFound
	}

	return &responseImpl{
		status: status,
		header: header,
	}
}

// Text sends a text response
func Text(status int, text string) Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, ctPlain)
	return &responseImpl{
		status: status,
		header: header,
		value:  text,
	}
}

// HTML creates a HTML response
func HTML(status int, html string) Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, ctHTML)
	return &responseImpl{
		status: status,
		header: header,
		value:  html,
	}
}

// JSON creates a application/json response
func JSON(status int, value interface{}) Responsible {
	header := make(http.Header)
	header.Set(mime.ContentType, ctJSON)
	return &responseImpl{
		status: status,
		header: header,
		value:  value,
	}
}

// Stream creates a application/octet-stream response
func Stream(r io.Reader) Responsible {
	return ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(mime.ContentType, mime.OctetStream)
		const size = 1024
		buf := make([]byte, size)
		for {
			n, err := r.Read(buf)
			if err != nil {
				log.FromContext(ctx).Errorf("Read: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if n < size {
				break
			}
			gox.WriteAll(w, buf[:n])
		}
	})
}

// StreamData creates a application/octet-stream response
func StreamData(b []byte) Responsible {
	return ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(mime.ContentType, mime.OctetStream)
		err := gox.WriteAll(w, b)
		if err != nil {
			log.FromContext(ctx).Errorf("WriteAll: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// File creates a file response
func File(req *http.Request, filePath string) Responsible {
	return ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
		http.ServeFile(w, req, filePath)
	})
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func TemplateHTML(templates []*template.Template, templateName string, params interface{}) Responsible {
	return ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
		for _, tmpl := range templates {
			var err error
			if len(templateName) == 0 {
				err = tmpl.Execute(w, params)
			} else {
				err = tmpl.ExecuteTemplate(w, templateName, params)
			}

			if err == nil {
				break
			}
		}
	})
}

// Handle handles request with h
func Handle(req *http.Request, h http.Handler) Responsible {
	return ResponsibleFunc(func(ctx context.Context, w http.ResponseWriter) {
		h.ServeHTTP(w, req)
	})
}
