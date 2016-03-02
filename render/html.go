package render

import (
	"github.com/justintan/xtypes"
	"html/template"
	"net/http"
)

func TemplateHTML(writer http.ResponseWriter, tpl *template.Template, name string, params xtypes.M) error {
	writer.Header()["Content-Type"] = []string{xtypes.MIMEHTML + "; charset=utf-8"}
	if len(name) == 0 {
		return tpl.Execute(writer, params)
	}

	return tpl.ExecuteTemplate(writer, name, params)
}

func HTML(writer http.ResponseWriter, htmlText string) {
	writer.Header()["Content-Type"] = []string{xtypes.MIMEHTML + "; charset=utf-8"}
	writer.Write([]byte(htmlText))
}
