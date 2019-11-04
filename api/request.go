package api

import (
	"encoding/json"

	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

// ParseRequest unmarshal request body into params. Only support JSON type for now.
func ParseParams(req *wine.Request, params interface{}) error {
	if req.ContentType() != mime.JSON {
		return gox.BadRequest("Expected %s instead of %s", mime.JSON, req.ContentType())
	}
	return json.Unmarshal(req.Body(), params)
}
