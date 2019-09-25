package api

import (
	"encoding/json"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

// ParseRequest unmarshal request body into params. Only support JSON type for now.
func ParseRequest(req *wine.Request, params interface{}) error {
	if req.ContentType() != mime.JSON {
		return nil
	}
	return json.Unmarshal(req.Body(), params)
}
