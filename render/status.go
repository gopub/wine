package render

import (
	"net/http"
)

// Status writes status code
func Status(writer http.ResponseWriter, status int) {
	http.Error(writer, http.StatusText(status), status)
}
