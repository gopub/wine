package wine

import (
	"github.com/justintan/gox"
	"net/http"
)

type NewContextFunc func(writer http.ResponseWriter, req *http.Request, handlers []Handler) Context

type Context interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Written() bool
	MarkWritten()
	HandlerChain() *HandlerChain
	Next()
	Request() *http.Request
	RequestParams() gox.M
	RequestHeader() http.Header
	ResponseHeader() http.Header
	ResponseWriter() http.ResponseWriter
	SendJSON(obj interface{})
	SendStatus(status int)
	SendFile(filePath string)
}
