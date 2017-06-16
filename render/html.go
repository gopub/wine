package render

import (
	"html/template"
	"log"
	"net/http"

	ghttp "github.com/natande/gox/http"
	gio "github.com/natande/gox/io"
)

// TemplateHTML render tmpl with name, params and writes into writer
func TemplateHTML(writer http.ResponseWriter, tmpl *template.Template, name string, params interface{}) error {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	if len(name) == 0 {
		return tmpl.Execute(writer, params)
	}

	return tmpl.ExecuteTemplate(writer, name, params)
}

// HTML writes htmlText into writer
func HTML(writer http.ResponseWriter, htmlText string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	var data = []byte(htmlText)
	err := gio.Write(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
