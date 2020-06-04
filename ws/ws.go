package ws

import (
	"context"

	"github.com/gopub/log"
	"github.com/gopub/wine/errors"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

type contextKey int

// Context keys
const (
	ckNextHandler contextKey = iota + 1
)

func Next(ctx context.Context, req *Request) *Response {
	i, _ := ctx.Value(ckNextHandler).(Handler)
	if i == nil {
		return &Response{
			ID:    req.ID,
			Error: errors.NotImplemented(""),
		}
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ckNextHandler, h)
}
