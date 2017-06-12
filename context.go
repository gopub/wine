package wine

import (
	"html/template"
	"net/http"

	"github.com/justintan/gox/types"
)

type Encoding string

const (
	EncodingNone    = Encoding("")
	EncodingGzip    = Encoding("gzip")
	EncodingDefalte = Encoding("deflate")
)

// Context defines a request context
type Context interface {
	HTTPRequest() *http.Request
	Params() types.M
	Header() http.Header

	Next()
	Responded() bool
	ResponseHeader() http.Header
	Status(status int)
	Redirect(location string, permanent bool)
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
