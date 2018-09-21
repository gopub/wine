package wine

import (
	"net/http"

	"github.com/gopub/types"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request struct {
	HTTPRequest *http.Request
	Parameters  types.M
}
