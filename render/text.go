package render

import (
	"net/http"

	"log"

	ghttp "github.com/justintan/gox/http"
	gio "github.com/justintan/gox/io"
)

func Text(writer http.ResponseWriter, text string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMETEXT + "; charset=utf-8"}
	data := []byte(text)

	writer.WriteHeader(http.StatusOK)
	err := gio.Write(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
