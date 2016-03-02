package render

import (
	"encoding/json"
	"github.com/justintan/xtypes"
	"net/http"
)

func JSON(writer http.ResponseWriter, jsonObj interface{}) {
	writer.Header()["Content-Type"] = []string{xtypes.MIMEJSON + "; charset=utf-8"}
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	writer.Write(jsonBytes)
}
