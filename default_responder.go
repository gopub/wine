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

// Reset can construct Context object with new parameters in order to make it reusable
func (c *DefaultResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	c.responded = false
	c.req = req
	c.writer = rw
	c.templates = tmpls
}

// Header returns response header
func (c *DefaultResponder) Header() http.Header {
	return c.writer.Header()
}

// Responded returns a flag to determine whether if the response has been written
func (c *DefaultResponder) Responded() bool {
	return c.responded
}

func (c *DefaultResponder) markResponded() {
	if c.responded {
		panic("[WINE] already responded")
	}
	c.responded = true
}

// Send sends bytes
func (c *DefaultResponder) Send(data []byte, contentType string) {
	c.markResponded()
	if len(contentType) == 0 {
		contentType = http.DetectContentType(data)
	}
	if strings.Index(contentType, "charset") < 0 {
		contentType += "; charset=utf-8"
	}
	c.Header()["Content-Type"] = []string{contentType}
	err := gox.WriteAll(c.writer, data)
	if err != nil {
		log.Println("[WINE] Send error:", err)
	}
}

// JSON sends json response
func (c *DefaultResponder) JSON(jsonObj interface{}) {
	c.markResponded()
	render.JSON(c.writer, jsonObj)
}

// Status sends a response just with a status code
func (c *DefaultResponder) Status(status int) {
	c.markResponded()
	render.Status(c.writer, status)
}

// Redirect sends a redirect response
func (c *DefaultResponder) Redirect(location string, permanent bool) {
	c.writer.Header().Set("Location", location)
	if permanent {
		c.Status(http.StatusMovedPermanently)
	} else {
		c.Status(http.StatusFound)
	}
}

// File sends a file response
func (c *DefaultResponder) File(filePath string) {
	c.markResponded()
	http.ServeFile(c.writer, c.req, filePath)
}

// HTML sends a HTML response
func (c *DefaultResponder) HTML(htmlText string) {
	c.markResponded()
	render.HTML(c.writer, htmlText)
}

// Text sends a text response
func (c *DefaultResponder) Text(text string) {
	c.markResponded()
	render.Text(c.writer, text)
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (c *DefaultResponder) TemplateHTML(templateName string, params interface{}) {
	for _, tmpl := range c.templates {
		err := render.TemplateHTML(c.writer, tmpl, templateName, params)
		if err == nil {
			c.markResponded()
			break
		}
	}
}

// Handle handles request with h
func (c *DefaultResponder) Handle(h http.Handler) {
	c.markResponded()
	h.ServeHTTP(c.writer, c.req)
}
