package wine

import (
	"net/http"

	"github.com/gopub/types"
)

// Request is a wrapper of http.Request, aims to provide more convenient interface
type Request interface {
	SetValue(key string, value interface{})
	Value(key string) interface{}
	RawRequest() *http.Request
	Parameters() types.M
}
