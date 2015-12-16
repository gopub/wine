package wine

import (
	"github.com/justintan/gox"
	"html/template"
	"net/http"
)

type Context interface {
	HTTPRequest() *http.Request
	Params() gox.M
	Header() http.Header

	Next()
	Responded() bool
	ResponseHeader() http.Header
	Status(status int)
	JSON(obj interface{})
	File(filePath string)
	HTML(htmlText string)
	TemplateHTML(templateName string, params gox.M)
	ServeHTTP(h http.Handler)

	Set(key string, value interface{})
	Get(key string) interface{}

	Rebuild(http.ResponseWriter, *http.Request, []*template.Template, []Handler)
}
