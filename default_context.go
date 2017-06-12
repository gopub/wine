package wine

import (
	"html/template"
	"net/http"
	"strings"

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
	encoding  Encoding

	req        *http.Request
	reqHeader  http.Header
	reqParams  types.M
	respHeader http.Header
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

	if dc.reqHeader != nil {
		for k := range dc.reqHeader {
			delete(dc.reqHeader, k)
		}
	} else {
		dc.reqHeader = make(http.Header)
	}

	if dc.respHeader != nil {
		for k := range dc.respHeader {
			delete(dc.respHeader, k)
		}
	} else {
		dc.respHeader = make(http.Header)
	}

	dc.responded = false
	dc.writer = rw
	dc.req = req
	dc.reqParams = ghttp.ParseParameters(req)
	dc.handlers = NewHandlerChain(handlers)
	dc.templates = templates
	dc.encoding = EncodingGzip

	for k, v := range req.Header {
		if strings.Index(k, "x-") == 0 {
			k = k[2:]
		}
		dc.reqHeader[k] = v
	}
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

// HTTPRequest returns request
func (dc *DefaultContext) HTTPRequest() *http.Request {
	return dc.req
}

// Params returns request's parameters including queries, body
func (dc *DefaultContext) Params() types.M {
	return dc.reqParams
}

// Header returns request header
func (dc *DefaultContext) Header() http.Header {
	return dc.reqHeader
}

// ResponseHeader returns response header
func (dc *DefaultContext) ResponseHeader() http.Header {
	return dc.respHeader
}

// Responded returns a flag to determine whether if the response has been written
func (dc *DefaultContext) Responded() bool {
	return dc.responded
}

func (dc *DefaultContext) setResponded() {
	if dc.responded {
		panic("[WINE] already responded")
	}
	dc.responded = true
}

func (dc *DefaultContext) parseCompression() string {
	encodings := dc.reqHeader.Get("Accept-Encoding")
	if strings.Index(encodings, "gzip") >= 0 {
		return "gzip"
	} else if strings.Index(encodings, "defalte") >= 0 {
		return "defalte"
	} else {
		return ""
	}
}

// JSON sends json response
func (dc *DefaultContext) JSON(jsonObj interface{}) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.JSON(dc.writer, jsonObj, dc.parseCompression())
}

// Status sends a response just with a status code
func (dc *DefaultContext) Status(status int) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.Status(dc.writer, status)
}

// Redirect sends a redirect response
func (dc *DefaultContext) Redirect(location string, permanent bool) {
	dc.respHeader.Set("Location", location)
	if permanent {
		dc.Status(http.StatusMovedPermanently)
	} else {
		dc.Status(http.StatusFound)
	}
}

// File sends a file response
func (dc *DefaultContext) File(filePath string) {
	dc.setResponded()
	http.ServeFile(dc.writer, dc.req, filePath)
}

// HTML sends a HTML response
func (dc *DefaultContext) HTML(htmlText string) {
	dc.setResponded()
	render.HTML(dc.writer, htmlText, dc.parseCompression())
}

// Text sends a text response
func (dc *DefaultContext) Text(text string) {
	dc.setResponded()
	render.Text(dc.writer, text, dc.parseCompression())
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (dc *DefaultContext) TemplateHTML(templateName string, params interface{}) {
	for _, tpl := range dc.templates {
		err := render.TemplateHTML(dc.writer, tpl, templateName, params)
		if err == nil {
			dc.setResponded()
			break
		}
	}
}

// ServeHTTP starts http server
func (dc *DefaultContext) ServeHTTP(h http.Handler) {
	dc.setResponded()
	h.ServeHTTP(dc.writer, dc.req)
}
