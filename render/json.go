package render

import (
	"encoding/json"
	"github.com/gopub/log"
	"net/http"

	"github.com/gopub/utils"
)

// JSON responds application/json content
func JSON(writer http.ResponseWriter, status int, jsonObj interface{}) {
	writer.Header()["Content-Type"] = []string{utils.MIMEJSON + "; charset=utf-8"}
	data, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		log.Error("Render error:", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(status)
	err = utils.WriteAll(writer, data)
	if err != nil {
		log.Error("Render error:", err)
	}
}
