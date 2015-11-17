package wine

import (
	"github.com/justintan/gox"
	"html/template"
	"net/http"
)

type NewContextFunc func(writer http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) Context

type Context interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	Written() bool
	MarkWritten()
	HandlerChain() *handlerChain
	Templates() []*template.Template
	Next()
	HttpRequest() *http.Request
	RequestParams() gox.M
	RequestHeader() http.Header
	ResponseHeader() http.Header
	ResponseWriter() http.ResponseWriter
	SendJSON(obj interface{})
	SendStatus(status int)
	SendFile(filePath string)
	SendHTML(htmlText string)
	SendTemplateHTML(templateFileName string, params gox.M)
}
