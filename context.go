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
	Send(data []byte, contentType string)
	Status(status int)
	JSON(obj interface{})
	File(filePath string)
	HTML(htmlText string)
	Text(text string)
	TemplateHTML(templateName string, params interface{})
	Handle(h http.Handler)

	Set(key string, value interface{})
	Get(key string) interface{}

	Rebuild(
		w http.ResponseWriter,
		req *http.Request,
		templates []*template.Template,
		handlers []Handler,
		maxMemory int64,
	)
}
