package storage

import (
	"context"
	"strings"

	"github.com/gopub/wine"
)

type FileReader struct {
	r Reader
}

func (r *FileReader) HandleRequest(ctx context.Context, req *wine.Request) wine.Responder {
	segments := strings.Split(req.Request().URL.Path, "/")
	var name string
	for i := len(segments) - 1; i >= 0; i-- {
		name = segments[i]
		if name != "" {
			break
		}
	}
	data, err := r.r.Read(ctx, name)
	if err != nil {
		return wine.Error(err)
	}
	return wine.File(data, name)
}
