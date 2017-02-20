package render

import (
	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
	"net/http"
)

func Text(writer http.ResponseWriter, text string, gzipFlag bool) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMETEXT + "; charset=utf-8"}
	data := []byte(text)
	var err error
	if gzipFlag {
		data, err = gcompress.Gzip(data)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Encoding", "gzip")
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}
