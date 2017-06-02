package render

import (
	"encoding/json"
	"net/http"

	gcompress "github.com/justintan/gox/compress"
	ghttp "github.com/justintan/gox/http"
)

// JSON responds application/json content
func JSON(writer http.ResponseWriter, jsonObj interface{}, gzipFlag bool) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEJSON + "; charset=utf-8"}
	data, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
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
