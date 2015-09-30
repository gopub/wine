package render

import (
	"net/http"
)

func Status(writer http.ResponseWriter, status int) {
	http.Error(writer, http.StatusText(status), status)
}
