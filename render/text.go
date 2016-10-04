package render

import (
	ghttp "github.com/justintan/gox/http"
	"net/http"
)

func Text(writer http.ResponseWriter, text string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMETEXT + "; charset=utf-8"}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(text))
}
