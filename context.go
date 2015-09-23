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

	Request *http.Request
	Header  gox.M
	Params  gox.M
}

func NewContext(rw http.ResponseWriter, req *http.Request, handlers []Handler, params map[string]string) *Context {
	c := &Context{}
	c.writer = rw
	c.Request = req
	c.Params = gox.ParseHttpRequest(req)
	if len(params) > 0 {
		c.Params.AddMap(params)
	}
	c.Header = gox.M{}
	for k, v := range req.Header {
		c.Header[strings.ToLower(k)] = v
	}
	c.handlers = handlers
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
	this.writer.Header().Set("Access-Control-Allow-Origin", "*")
	this.writer.Header()[gox.ContentTypeName] = gox.JsonContentType
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
	this.writer.Header().Set("Access-Control-Allow-Origin", "*")
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
