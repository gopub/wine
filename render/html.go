package render

import (
	"github.com/gopub/log"
	"html/template"
	"net/http"

	"github.com/gopub/utils"
)

// TemplateHTML render tmpl with name, params and writes into writer
func TemplateHTML(writer http.ResponseWriter, tmpl *template.Template, name string, params interface{}) error {
	writer.Header()["Content-Type"] = []string{utils.MIMEHTML + "; charset=utf-8"}
	if len(name) == 0 {
		return tmpl.Execute(writer, params)
	}

	return tmpl.ExecuteTemplate(writer, name, params)
}

// HTML writes htmlText into writer
func HTML(writer http.ResponseWriter, htmlText string) {
	writer.Header()["Content-Type"] = []string{utils.MIMEHTML + "; charset=utf-8"}
	var data = []byte(htmlText)
	err := utils.WriteAll(writer, data)
	if err != nil {
		log.Error("Render error:", err)
	}
}
