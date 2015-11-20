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
	c.writer = rw
	c.req = req
	c.reqParams = gox.ParseHttpRequest(req)
	c.reqHeader = make(http.Header)
	for k, v := range req.Header {
		c.reqHeader[strings.ToLower(k)] = v
	}
	c.handlers = NewHandlerChain(handlers)
	c.respHeader = make(http.Header)
	c.templates = templates
	return c
}

func (this *DefaultContext) Set(key string, value interface{}) {
	this.keyValues[key] = value
}

func (this *DefaultContext) Get(key string) interface{} {
	return this.keyValues[key]
}

func (this *DefaultContext) Next() {
	if h := this.handlers.Next(); h != nil {
		h.HandleRequest(this)
	}
}

func (this *DefaultContext) HTTPRequest() *http.Request {
	return this.req
}

func (this *DefaultContext) Params() gox.M {
	return this.reqParams
}

func (this *DefaultContext) Header() http.Header {
	return this.reqHeader
}

func (this *DefaultContext) ResponseHeader() http.Header {
	return this.respHeader
}

func (this *DefaultContext) Responded() bool {
	return this.responded
}

func (this *DefaultContext) setResponded() {
	if this.responded {
		panic("cannot responded twice")
	}
	this.responded = true
}

func (this *DefaultContext) JSON(jsonObj interface{}) {
	this.setResponded()
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	render.JSON(this.writer, jsonObj)
}

func (this *DefaultContext) Status(status int) {
	this.setResponded()
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	render.Status(this.writer, status)
}

func (this *DefaultContext) File(filePath string) {
	this.setResponded()
	http.ServeFile(this.writer, this.req, filePath)
}

func (this *DefaultContext) HTML(htmlText string) {
	this.setResponded()
	render.HTML(this.writer, htmlText)
}

func (this *DefaultContext) TemplateHTML(templateName string, params gox.M) {
	for _, tpl := range this.templates {
		err := render.TemplateHTML(this.writer, tpl, templateName, params)
		if err == nil {
			this.setResponded()
			break
		}
	}
}

func (this *DefaultContext) ServeHTTP(h http.Handler) {
	this.setResponded()
	h.ServeHTTP(this.writer, this.req)
}
