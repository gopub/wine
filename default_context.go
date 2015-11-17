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
	written   bool
	templates []*template.Template
	handlers  *handlerChain
	index     int //handler index

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

func (this *DefaultContext) Written() bool {
	return this.written
}

func (this *DefaultContext) MarkWritten() {
	this.written = true
}

func (this *DefaultContext) HandlerChain() *handlerChain {
	return this.handlers
}

func (this *DefaultContext) Next() {
	if h := this.handlers.Next(); h != nil {
		h.HandleRequest(this)
	}
}

func (this *DefaultContext) HttpRequest() *http.Request {
	return this.req
}

func (this *DefaultContext) RequestParams() gox.M {
	return this.reqParams
}

func (this *DefaultContext) RequestHeader() http.Header {
	return this.reqHeader
}

func (this *DefaultContext) ResponseHeader() http.Header {
	return this.respHeader
}

func (this *DefaultContext) ResponseWriter() http.ResponseWriter {
	return this.writer
}

func (this *DefaultContext) Templates() []*template.Template {
	return this.templates
}

func (this *DefaultContext) SendJSON(jsonObj interface{}) {
	if this.written {
		panic("already written")
	}
	this.written = true
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	render.JSON(this.writer, jsonObj)
}

func (this *DefaultContext) SendStatus(status int) {
	if this.written {
		panic("already written")
	}
	this.written = true
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	render.Status(this.writer, status)
}

func (this *DefaultContext) SendFile(filePath string) {
	http.ServeFile(this.ResponseWriter(), this.HttpRequest(), filePath)
}

func (this *DefaultContext) SendHTML(htmlText string) {
	this.MarkWritten()
	render.HTML(this.ResponseWriter(), htmlText)
}

func (this *DefaultContext) SendTemplateHTML(templateFileName string, params gox.M) {
	for _, tpl := range this.templates {
		err := render.TemplateHTML(this.ResponseWriter(), tpl, templateFileName, params)
		if err == nil {
			this.MarkWritten()
			break
		}
	}
}
