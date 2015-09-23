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

	Request *http.Request
	Header  gox.M
	Params  gox.M
}

func NewContext(rw http.ResponseWriter, req *http.Request) *Context {
	c := &Context{}
	c.writer = rw
	c.Request = req
	c.Params = gox.ParseHttpRequest(req)
	c.Header = gox.M{}
	for k, v := range req.Header {
		c.Header[strings.ToLower(k)] = v
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
	this.writer.WriteHeader(s)
}
