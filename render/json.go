package render

import (
	"encoding/json"
	"github.com/justintan/gox"
	"net/http"
)

func JSON(writer http.ResponseWriter, jsonObj interface{}) {
	writer.Header()[gox.ContentTypeName] = gox.JSONContentType
	jsonBytes, err := json.MarshalIndent(jsonObj, "", "    ")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
	writer.Write(jsonBytes)
}
