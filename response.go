package wine

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

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

// Responder interface is used by Wine server to write response to the client
type Responder interface {
	// Respond will be called to write status/body to http response writer
	Respond(ctx context.Context, w http.ResponseWriter)
}

// ResponderFunc is a func that implements interface Responder
type ResponderFunc func(ctx context.Context, w http.ResponseWriter)

func (f ResponderFunc) Respond(ctx context.Context, w http.ResponseWriter) {
	f(ctx, w)
}

// Response holds all the http response information
// Value and headers except the status code can be modified before sent to the client
type Response struct {
	Responder
	status int
	header http.Header
	value  interface{}
}

// Respond writes header and body to response writer w
func (r *Response) Respond(ctx context.Context, w http.ResponseWriter) {
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

// Status returns a response only with a status code
func Status(status int) Responder {
	return Text(status, http.StatusText(status))
}

// Redirect sends a redirect response
func Redirect(location string, permanent bool) Responder {
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
func Text(status int, text string) Responder {
	header := make(http.Header)
	header.Set(mime.ContentType, ctPlain)
	return &Response{
		status: status,
		header: header,
		value:  text,
	}
}

// HTML creates a HTML response
func HTML(status int, html string) Responder {
	header := make(http.Header)
	header.Set(mime.ContentType, ctHTML)
	return &Response{
		status: status,
		header: header,
		value:  html,
	}
}

// JSON creates a application/json response
func JSON(status int, value interface{}) Responder {
	header := make(http.Header)
	header.Set(mime.ContentType, ctJSON)
	return &Response{
		status: status,
		header: header,
		value:  value,
	}
}

// StreamFile creates a application/octet-stream response
func StreamFile(r io.Reader, name string) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		logger := log.FromContext(ctx)
		w.Header().Set(mime.ContentType, mime.OctetStream)
		if name != "" {
			w.Header().Set(mime.ContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, name))
		}
		const size = 1024
		buf := make([]byte, size)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				if _, wErr := w.Write(buf[:n]); wErr != nil {
					logger.Errorf("Write: %v", wErr)
					return
				}
			}
			if err != nil {
				logger.Errorf("Read: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})
}

// File creates a application/octet-stream response
func File(b []byte, name string) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(mime.ContentType, mime.OctetStream)
		if name != "" {
			w.Header().Set(mime.ContentDisposition, fmt.Sprintf(`attachment; filename="%s"`, name))
		}
		if _, err := w.Write(b); err != nil {
			log.FromContext(ctx).Errorf("write: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// StaticFile serves static files
func StaticFile(req *http.Request, filePath string) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		http.ServeFile(w, req, filePath)
	})
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func TemplateHTML(templates []*template.Template, templateName string, params interface{}) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
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
func Handle(req *http.Request, h http.Handler) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		h.ServeHTTP(w, req)
	})
}
