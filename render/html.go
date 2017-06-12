package render

import (
	"html/template"
	"log"
	"net/http"

	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
	gio "github.com/justintan/gox/io"
)

func TemplateHTML(writer http.ResponseWriter, tmpl *template.Template, name string, params interface{}) error {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	if len(name) == 0 {
		return tmpl.Execute(writer, params)
	}

	return tmpl.ExecuteTemplate(writer, name, params)
}

func HTML(writer http.ResponseWriter, htmlText string, compression string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEHTML + "; charset=utf-8"}
	var data = []byte(htmlText)
	if len(compression) > 0 {
		compressor := gcompress.GetCompressor(compression)
		if compressor == nil {
			log.Println("[WINE] No compressor for:", compression)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		var err error
		data, err = compressor.Compress(data)
		if err != nil {
			log.Println("[WINE] Compression error:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Encoding", compression)
	}
	writer.WriteHeader(http.StatusOK)
	err := gio.Write(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
