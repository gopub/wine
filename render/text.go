package render

import (
	"github.com/gopub/log"
	"github.com/gopub/utils"
	"net/http"
)

// Text render text into writer
func Text(writer http.ResponseWriter, status int, text string) {
	writer.Header()["Content-Type"] = []string{utils.MIMETEXT + "; charset=utf-8"}
	writer.WriteHeader(status)
	data := []byte(text)
	err := utils.WriteAll(writer, data)
	if err != nil {
		log.Error("Render error:", err)
	}
}
