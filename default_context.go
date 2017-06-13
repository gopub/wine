package wine

import (
	"html/template"
	"net/http"

	ghttp "github.com/justintan/gox/http"
	"github.com/justintan/gox/types"
	"github.com/justintan/wine/render"
)

// DefaultContext is a default implementation of Context interface
type DefaultContext struct {
	keyValues types.M
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
	handlers  *HandlerChain

	req       *http.Request
	reqParams types.M
}

// Rebuild can construct Context object with new parameters in order to make it reusable
func (dc *DefaultContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) {
	if dc.keyValues != nil {
		for k := range dc.keyValues {
			delete(dc.keyValues, k)
		}
	} else {
		dc.keyValues = types.M{}
	}

	dc.responded = false
	dc.writer = rw
	dc.req = req
	dc.reqParams = ghttp.ParseParameters(req)
	dc.handlers = NewHandlerChain(handlers)
	dc.templates = templates
}

// Set sets key:value
func (dc *DefaultContext) Set(key string, value interface{}) {
	dc.keyValues[key] = value
}

// Get returns value for key
func (dc *DefaultContext) Get(key string) interface{} {
	return dc.keyValues[key]
}

// Next calls the next handler
func (dc *DefaultContext) Next() {
	if h := dc.handlers.Next(); h != nil {
		h.HandleRequest(dc)
	}
}

// Request returns request
func (dc *DefaultContext) Request() *http.Request {
	return dc.req
}

// Params returns request's parameters including queries, body
func (dc *DefaultContext) Params() types.M {
	return dc.reqParams
}

// Header returns response header
func (dc *DefaultContext) Header() http.Header {
	return dc.writer.Header()
}

// Responded returns a flag to determine whether if the response has been written
func (dc *DefaultContext) Responded() bool {
	return dc.responded
}

func (dc *DefaultContext) markResponded() {
	if dc.responded {
		panic("[WINE] already responded")
	}
	dc.responded = true
}

// JSON sends json response
func (dc *DefaultContext) JSON(jsonObj interface{}) {
	dc.markResponded()
	render.JSON(dc.writer, jsonObj)
}

// Status sends a response just with a status code
func (dc *DefaultContext) Status(status int) {
	dc.markResponded()
	render.Status(dc.writer, status)
}

// Redirect sends a redirect response
func (dc *DefaultContext) Redirect(location string, permanent bool) {
	dc.writer.Header().Set("Location", location)
	if permanent {
		dc.Status(http.StatusMovedPermanently)
	} else {
		dc.Status(http.StatusFound)
	}
}

// File sends a file response
func (dc *DefaultContext) File(filePath string) {
	dc.markResponded()
	http.ServeFile(dc.writer, dc.req, filePath)
}

// HTML sends a HTML response
func (dc *DefaultContext) HTML(htmlText string) {
	dc.markResponded()
	render.HTML(dc.writer, htmlText)
}

// Text sends a text response
func (dc *DefaultContext) Text(text string) {
	dc.markResponded()
	render.Text(dc.writer, text)
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (dc *DefaultContext) TemplateHTML(templateName string, params interface{}) {
	for _, tmpl := range dc.templates {
		err := render.TemplateHTML(dc.writer, tmpl, templateName, params)
		if err == nil {
			dc.markResponded()
			break
		}
	}
}

// ServeHTTP handles request with h
func (dc *DefaultContext) ServeHTTP(h http.Handler) {
	dc.markResponded()
	h.ServeHTTP(dc.writer, dc.req)
}
