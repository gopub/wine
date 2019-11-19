package api

import (
	"encoding/json"
	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
)

// ParseParams unmarshal request body into params. Only support JSON type for now.
func ParseParams(req *wine.Request, params interface{}) error {
	// Unsafe assignment, so ignore error
	data, err := json.Marshal(req.Params())
	if err == nil {
		_ = json.Unmarshal(data, params)
	}
	_ = gox.CopyWithNamer(params, req.Params(), gox.SnakeToCamelNamer)

	if req.ContentType() == mime.JSON {
		err := json.Unmarshal(req.Body(), params)
		if err != nil {
			return gox.BadRequest("unmarshal: %v", err)
		}
	}
	return nil
}
