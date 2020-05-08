package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gopub/log"

	"github.com/gopub/types"
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

	if req.ContentType() == mime.JSON {
		err := json.Unmarshal(req.Body(), params)
		if err != nil {
			return types.NewError(http.StatusBadRequest, "unmarshal: %v", err)
		}
	}
	return nil
}

func StructToQuery(s interface{}) url.Values {
	q := url.Values{}
	m := types.M{}
	err := m.AddStruct(s)
	if err != nil {
		log.Error(err)
		return q
	}
	for k, v := range m {
		q.Set(k, fmt.Sprint(v))
	}
	return q
}
