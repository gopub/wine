package api

import (
	"encoding/json"

	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

// ParseRequest unmarshal request body into params. Only support JSON type for now.
func ParseParams(req *wine.Request, params interface{}) error {
	if req.ContentType() == mime.JSON {
		err := json.Unmarshal(req.Body(), params)
		if err != nil {
			return gox.BadRequest("unmarshal: %v", err)
		}
		return nil
	}
	err := gox.CopyWithNamer(params, req.Params(), gox.SnakeToCamelNamer)
	if err != nil {
		return gox.BadRequest("parse params: %v", err)
	}
	return nil
}
