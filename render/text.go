package render

import (
	"log"
	"net/http"

	"github.com/natande/gox"
)

// Text render text into writer
func Text(writer http.ResponseWriter, text string) {
	writer.Header()["Content-Type"] = []string{gox.MIMETEXT + "; charset=utf-8"}
	data := []byte(text)
	err := gox.WriteAll(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
