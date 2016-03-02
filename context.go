package wine

import (
	"github.com/justintan/xtypes"
	"html/template"
	"net/http"
)

type Context interface {
	HTTPRequest() *http.Request
	Params() xtypes.M
	Header() http.Header

	Next()
	Responded() bool
	ResponseHeader() http.Header
	Status(status int)
	JSON(obj interface{})
	File(filePath string)
	HTML(htmlText string)
	TemplateHTML(templateName string, params xtypes.M)
	ServeHTTP(h http.Handler)

	Set(key string, value interface{})
	Get(key string) interface{}

	Rebuild(http.ResponseWriter, *http.Request, []*template.Template, []Handler)
}
