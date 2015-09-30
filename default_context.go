package wine

import (
	"github.com/justintan/gox"
	"github.com/justintan/wine/render"
	"net/http"
	"strings"
)

type DefaultContext struct {
	keyValues gox.M
	writer    http.ResponseWriter
	written   bool
	handlers  *HandlerChain
	index     int //handler index

	req        *http.Request
	reqHeader  http.Header
	reqParams  gox.M
	respHeader http.Header
}

func NewDefaultContext(rw http.ResponseWriter, req *http.Request, handlers []Handler) Context {
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

func (this *DefaultContext) HandlerChain() *HandlerChain {
	return this.handlers
}

func (this *DefaultContext) Next() {
	if h := this.handlers.Next(); h != nil {
		h(this)
	}
}

func (this *DefaultContext) Request() *http.Request {
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
