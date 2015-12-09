package render

import (
	"github.com/justintan/gox"
	"html/template"
	"net/http"
)

func TemplateHTML(writer http.ResponseWriter, tpl *template.Template, name string, params gox.M) error {
	writer.Header()[gox.ContentTypeName] = gox.HTMLContentType
	if len(name) == 0 {
		return tpl.Execute(writer, params)
	}

	return tpl.ExecuteTemplate(writer, name, params)
}

func HTML(writer http.ResponseWriter, htmlText string) {
	writer.Header()[gox.ContentTypeName] = gox.HTMLContentType
	writer.Write([]byte(htmlText))
}
