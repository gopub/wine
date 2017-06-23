package wine

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/natande/gox"
	"github.com/natande/wine/render"
)

var _ Context = (*DefaultContext)(nil)

// DefaultContext is a default implementation of Context interface
type DefaultContext struct {
	keyValues gox.M
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
	handlers  *HandlerChain

	req       *http.Request
	reqParams gox.M
}

// Rebuild can construct Context object with new parameters in order to make it reusable
func (c *DefaultContext) Rebuild(
	rw http.ResponseWriter,
	req *http.Request,
	templates []*template.Template,
	handlers []Handler,
	maxMemory int64,
) {
	if c.keyValues != nil {
		for k := range c.keyValues {
			delete(c.keyValues, k)
		}
	} else {
		c.keyValues = gox.M{}
	}

	c.responded = false
	c.writer = rw
	c.req = req
	c.reqParams = gox.ParseHTTPRequestParameters(req, maxMemory)
	c.handlers = NewHandlerChain(handlers)
	c.templates = templates
}

// Set sets key:value
func (c *DefaultContext) Set(key string, value interface{}) {
	c.keyValues[key] = value
}

// Get returns value for key
func (c *DefaultContext) Get(key string) interface{} {
	return c.keyValues[key]
}

// Next calls the next handler
func (c *DefaultContext) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

// Request returns request
func (c *DefaultContext) Request() *http.Request {
	return c.req
}

// Params returns request's parameters including queries, body
func (c *DefaultContext) Params() gox.M {
	return c.reqParams
}

// Header returns response header
func (c *DefaultContext) Header() http.Header {
	return c.writer.Header()
}

// Responded returns a flag to determine whether if the response has been written
func (c *DefaultContext) Responded() bool {
	return c.responded
}

func (c *DefaultContext) markResponded() {
	if c.responded {
		panic("[WINE] already responded")
	}
	c.responded = true
}

// Send sends bytes
func (c *DefaultContext) Send(data []byte, contentType string) {
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
func (c *DefaultContext) JSON(jsonObj interface{}) {
	c.markResponded()
	render.JSON(c.writer, jsonObj)
}

// Status sends a response just with a status code
func (c *DefaultContext) Status(status int) {
	c.markResponded()
	render.Status(c.writer, status)
}

// Redirect sends a redirect response
func (c *DefaultContext) Redirect(location string, permanent bool) {
	c.writer.Header().Set("Location", location)
	if permanent {
		c.Status(http.StatusMovedPermanently)
	} else {
		c.Status(http.StatusFound)
	}
}

// File sends a file response
func (c *DefaultContext) File(filePath string) {
	c.markResponded()
	http.ServeFile(c.writer, c.req, filePath)
}

// HTML sends a HTML response
func (c *DefaultContext) HTML(htmlText string) {
	c.markResponded()
	render.HTML(c.writer, htmlText)
}

// Text sends a text response
func (c *DefaultContext) Text(text string) {
	c.markResponded()
	render.Text(c.writer, text)
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (c *DefaultContext) TemplateHTML(templateName string, params interface{}) {
	for _, tmpl := range c.templates {
		err := render.TemplateHTML(c.writer, tmpl, templateName, params)
		if err == nil {
			c.markResponded()
			break
		}
	}
}

// Handle handles request with h
func (c *DefaultContext) Handle(h http.Handler) {
	c.markResponded()
	h.ServeHTTP(c.writer, c.req)
}
