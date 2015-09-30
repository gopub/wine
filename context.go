package wine

import (
	"encoding/json"
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type Context interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Written() bool
	SendJSON(obj interface{})
	SendStatus(status int)
	Next()
	Request() *http.Request
	RequestParams() gox.M
	RequestHeader() http.Header
	ResponseHeader() http.Header
	ResponseWriter() http.ResponseWriter
}

type DefaultContext struct {
	keyValues gox.M
	writer    http.ResponseWriter
	written   bool
	handlers  []Handler
	index     int //handler index

	req        *http.Request
	reqHeader  http.Header
	reqParams  gox.M
	respHeader http.Header
}

func NewContext(rw http.ResponseWriter, req *http.Request, handlers []Handler, params map[string]string, header http.Header) *DefaultContext {
	c := &DefaultContext{}
	c.keyValues = gox.M{}
	c.writer = rw
	c.req = req
	c.reqParams = gox.ParseHttpRequest(req)
	if len(params) > 0 {
		c.reqParams.AddMap(params)
	}
	c.reqHeader = make(http.Header)
	for k, v := range req.Header {
		c.reqHeader[strings.ToLower(k)] = v
	}
	c.handlers = handlers
	c.respHeader = make(http.Header)
	for k, v := range header {
		c.respHeader[k] = v
	}
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

func (this *DefaultContext) SendJSON(obj interface{}) {
	if this.written {
		panic("already written")
	}
	this.written = true
	this.writer.Header()[gox.ContentTypeName] = gox.JsonContentType
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	jsonBytes, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		this.writer.WriteHeader(http.StatusInternalServerError)
	} else {
		this.writer.WriteHeader(http.StatusOK)
	}
	this.writer.Write(jsonBytes)
}

func (this *DefaultContext) SendStatus(s int) {
	if this.written {
		panic("already written")
	}
	this.written = true
	for k, v := range this.respHeader {
		this.writer.Header()[k] = v
	}
	http.Error(this.writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (this *DefaultContext) Next() {
	if this.index >= len(this.handlers) {
		return
	}

	index := this.index
	this.index += 1
	this.handlers[index](this)
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
