package wine

import (
	"encoding/json"
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type Context struct {
	keyValues gox.M
	writer    http.ResponseWriter
	written   bool
	handlers  []Handler
	index     int //handler index

	Request        *http.Request
	RequestHeader  gox.M
	RequestParams  gox.M
	ResponseHeader http.Header
}

func NewContext(rw http.ResponseWriter, req *http.Request, handlers []Handler, params map[string]string, header http.Header) *Context {
	c := &Context{}
	c.writer = rw
	c.Request = req
	c.RequestParams = gox.ParseHttpRequest(req)
	if len(params) > 0 {
		c.RequestParams.AddMap(params)
	}
	c.RequestHeader = gox.M{}
	for k, v := range req.Header {
		c.RequestHeader[strings.ToLower(k)] = v
	}
	c.handlers = handlers
	c.ResponseHeader = make(http.Header)
	for k, v := range header {
		c.ResponseHeader[k] = v
	}
	return c
}

func (this *Context) Set(key string, value interface{}) {
	this.keyValues[key] = value
}

func (this *Context) Get(key string) interface{} {
	return this.keyValues[key]
}

func (this *Context) Written() bool {
	return this.written
}

func (this *Context) JSON(obj interface{}) {
	if this.written {
		panic("already written")
	}
	this.written = true
	this.writer.Header()[gox.ContentTypeName] = gox.JsonContentType
	for k, v := range this.ResponseHeader {
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

func (this *Context) Status(s int) {
	if this.written {
		panic("already written")
	}
	this.written = true
	for k, v := range this.ResponseHeader {
		this.writer.Header()[k] = v
	}
	http.Error(this.writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (this *Context) Next() {
	if this.index >= len(this.handlers) {
		return
	}

	index := this.index
	this.index += 1
	this.handlers[index](this)
}
