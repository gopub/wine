package render

import (
	"encoding/json"
	"log"
	"net/http"

	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
	gio "github.com/justintan/gox/io"
)

// JSON responds application/json content
func JSON(writer http.ResponseWriter, jsonObj interface{}, compression string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEJSON + "; charset=utf-8"}
	data, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		log.Println("[WINE] Render error:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(compression) > 0 {
		compressor := gcompress.GetCompressor(compression)
		if compressor == nil {
			log.Println("[WINE] Render error: no compressor for", compression)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		data, err = compressor.Compress(data)
		if err != nil {
			log.Println("[WINE] Render error:", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Encoding", compression)
	}
	writer.WriteHeader(http.StatusOK)
	err = gio.Write(writer, data)
	if err != nil {
		log.Println("[WINE] Render error:", err)
	}
}
