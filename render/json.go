package render

import (
	"encoding/json"
	"log"
	"net/http"

	ghttp "github.com/natande/gox/http"
	gio "github.com/natande/gox/io"
)

// JSON responds application/json content
func JSON(writer http.ResponseWriter, jsonObj interface{}) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEJSON + "; charset=utf-8"}
	data, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		log.Println("[WINE] Render error:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = gio.Write(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
