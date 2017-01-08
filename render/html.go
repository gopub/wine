package render

import (
	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
	"github.com/justintan/gox/types"
	"html/template"
	"net/http"
)

func TemplateHTML(writer http.ResponseWriter, tpl *template.Template, name string, params types.M) error {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	if len(name) == 0 {
		return tpl.Execute(writer, params)
	}

	return tpl.ExecuteTemplate(writer, name, params)
}

func HTML(writer http.ResponseWriter, htmlText string, gzipFlag bool) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	if gzipFlag {
		data, err := gcompress.Gzip([]byte(htmlText))
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Encoding", "gzip")
		writer.Write(data)
	} else {
		writer.Write([]byte(htmlText))
	}
}
