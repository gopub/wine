package wine

import (
	"github.com/justintan/gox"
	"github.com/justintan/wine/render"
	"html/template"
	"net/http"
	"strings"
)

type DefaultContext struct {
	keyValues gox.M
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
	handlers  *HandlerChain

	req        *http.Request
	reqHeader  http.Header
	reqParams  gox.M
	respHeader http.Header
}

func NewDefaultContext(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) Context {
	c := &DefaultContext{}
	c.keyValues = gox.M{}
	c.reqHeader = make(http.Header)
	c.respHeader = make(http.Header)
	c.Reborn(rw, req, templates, handlers)
	return c
}

func (dc *DefaultContext) Reborn(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) {
	for k := range dc.keyValues {
		delete(dc.keyValues, k)
	}
	for k := range dc.reqHeader {
		delete(dc.reqHeader, k)
	}
	for k := range dc.respHeader {
		delete(dc.respHeader, k)
	}
	dc.responded = false

	dc.writer = rw
	dc.req = req
	dc.reqParams = gox.ParseReqParams(req)
	dc.handlers = NewHandlerChain(handlers)
	dc.templates = templates

	for k, v := range req.Header {
		dc.reqHeader[strings.ToLower(k)] = v
	}
}

func (dc *DefaultContext) Set(key string, value interface{}) {
	dc.keyValues[key] = value
}

func (dc *DefaultContext) Get(key string) interface{} {
	return dc.keyValues[key]
}

func (dc *DefaultContext) Next() {
	if h := dc.handlers.Next(); h != nil {
		h.HandleRequest(dc)
	}
}

func (dc *DefaultContext) HTTPRequest() *http.Request {
	return dc.req
}

func (dc *DefaultContext) Params() gox.M {
	return dc.reqParams
}

func (dc *DefaultContext) Header() http.Header {
	return dc.reqHeader
}

func (dc *DefaultContext) ResponseHeader() http.Header {
	return dc.respHeader
}

func (dc *DefaultContext) Responded() bool {
	return dc.responded
}

func (dc *DefaultContext) setResponded() {
	if dc.responded {
		panic("cannot responded twice")
	}
	dc.responded = true
}

func (dc *DefaultContext) JSON(jsonObj interface{}) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.JSON(dc.writer, jsonObj)
}

func (dc *DefaultContext) Status(status int) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.Status(dc.writer, status)
}

func (dc *DefaultContext) File(filePath string) {
	dc.setResponded()
	http.ServeFile(dc.writer, dc.req, filePath)
}

func (dc *DefaultContext) HTML(htmlText string) {
	dc.setResponded()
	render.HTML(dc.writer, htmlText)
}

func (dc *DefaultContext) TemplateHTML(templateName string, params gox.M) {
	for _, tpl := range dc.templates {
		err := render.TemplateHTML(dc.writer, tpl, templateName, params)
		if err == nil {
			dc.setResponded()
			break
		}
	}
}

func (dc *DefaultContext) ServeHTTP(h http.Handler) {
	dc.setResponded()
	h.ServeHTTP(dc.writer, dc.req)
}
