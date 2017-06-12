package render

import (
	"net/http"

	"log"

	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
	gio "github.com/justintan/gox/io"
)

func Text(writer http.ResponseWriter, text string, compression string) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMETEXT + "; charset=utf-8"}
	data := []byte(text)
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
