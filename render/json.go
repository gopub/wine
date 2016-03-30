package render

import (
	"encoding/json"
	ghttp "github.com/justintan/gox/http"
	"net/http"
)

func JSON(writer http.ResponseWriter, jsonObj interface{}) {
	writer.Header()["Content-Type"] = []string{ghttp.MIMEJSON + "; charset=utf-8"}
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	writer.Write(jsonBytes)
}
