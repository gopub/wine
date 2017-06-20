package render

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/natande/gox"
)

// JSON responds application/json content
func JSON(writer http.ResponseWriter, jsonObj interface{}) {
	writer.Header()["Content-Type"] = []string{gox.MIMEJSON + "; charset=utf-8"}
	data, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		log.Println("[WINE] Render error:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = gox.WriteAll(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
