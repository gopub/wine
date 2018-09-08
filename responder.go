package wine

import (
	"html/template"
	"net/http"
)

// Responder defines methods to send response
type Responder interface {
	Responded() bool
	Header() http.Header
	Redirect(location string, permanent bool)
	Send(data []byte, contentType string)
	Status(status int)
	Text(status int, text string)
	HTML(status int, htmlText string)
	JSON(status int, obj interface{})
	File(filePath string)
	TemplateHTML(templateName string, params interface{})
	Handle(h http.Handler)

	Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template)
}
