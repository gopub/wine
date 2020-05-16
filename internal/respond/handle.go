package respond

import (
	"context"
	"net/http"
)

func Handle(req *http.Request, h http.Handler) Func {
	return func(ctx context.Context, w http.ResponseWriter) {
		h.ServeHTTP(w, req)
	}
}
