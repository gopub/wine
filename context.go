package wine

import (
	"html/template"
	"net/http"

	"github.com/justintan/gox/types"
)

// Context defines a request context
type Context interface {
	Request() *http.Request
	Params() types.M

	Next()
	Responded() bool
	Header() http.Header
	Redirect(location string, permanent bool)
	Status(status int)
	JSON(obj interface{})
	File(filePath string)
	HTML(htmlText string)
	Text(text string)
	TemplateHTML(templateName string, params interface{})
	ServeHTTP(h http.Handler)

	Set(key string, value interface{})
	Get(key string) interface{}

	Rebuild(http.ResponseWriter, *http.Request, []*template.Template, []Handler)
}
