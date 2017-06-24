package wine

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/natande/gox"
	"github.com/natande/wine/render"
)

var _ Responder = (*DefaultResponder)(nil)

// DefaultResponder is a default implementation of Context interface
type DefaultResponder struct {
	req       *http.Request
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
}

// Reset resets responder to be a new one
func (dr *DefaultResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	dr.responded = false
	dr.req = req
	dr.writer = rw
	dr.templates = tmpls
}

// Header returns response header
func (dr *DefaultResponder) Header() http.Header {
	return dr.writer.Header()
}

// Responded returns a flag to determine whether if the response has been written
func (dr *DefaultResponder) Responded() bool {
	return dr.responded
}

func (dr *DefaultResponder) markResponded() {
	if dr.responded {
		panic("[WINE] already responded")
	}
	dr.responded = true
}

// Send sends bytes
func (dr *DefaultResponder) Send(data []byte, contentType string) {
	dr.markResponded()
	if len(contentType) == 0 {
		contentType = http.DetectContentType(data)
	}
	if strings.Index(contentType, "charset") < 0 {
		contentType += "; charset=utf-8"
	}
	dr.Header()["Content-Type"] = []string{contentType}
	err := gox.WriteAll(dr.writer, data)
	if err != nil {
		log.Println("[WINE] Send error:", err)
	}
}

// JSON sends json response
func (dr *DefaultResponder) JSON(jsonObj interface{}) {
	dr.markResponded()
	render.JSON(dr.writer, jsonObj)
}

// Status sends a response just with a status code
func (dr *DefaultResponder) Status(status int) {
	dr.markResponded()
	render.Status(dr.writer, status)
}

// Redirect sends a redirect response
func (dr *DefaultResponder) Redirect(location string, permanent bool) {
	dr.writer.Header().Set("Location", location)
	if permanent {
		dr.Status(http.StatusMovedPermanently)
	} else {
		dr.Status(http.StatusFound)
	}
}

// File sends a file response
func (dr *DefaultResponder) File(filePath string) {
	dr.markResponded()
	http.ServeFile(dr.writer, dr.req, filePath)
}

// HTML sends a HTML response
func (dr *DefaultResponder) HTML(htmlText string) {
	dr.markResponded()
	render.HTML(dr.writer, htmlText)
}

// Text sends a text response
func (dr *DefaultResponder) Text(text string) {
	dr.markResponded()
	render.Text(dr.writer, text)
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (dr *DefaultResponder) TemplateHTML(templateName string, params interface{}) {
	for _, tmpl := range dr.templates {
		err := render.TemplateHTML(dr.writer, tmpl, templateName, params)
		if err == nil {
			dr.markResponded()
			break
		}
	}
}

// Handle handles request with h
func (dr *DefaultResponder) Handle(h http.Handler) {
	dr.markResponded()
	h.ServeHTTP(dr.writer, dr.req)
}
