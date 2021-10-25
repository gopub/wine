package wine

import (
	"context"
	"github.com/gopub/wine/router"

	"github.com/gopub/conv"
	"github.com/gopub/log/v2"
	"github.com/gopub/wine/ctxutil"
	"github.com/gopub/wine/internal/respond"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine")
	respond.SetLogger(logger)
	router.SetLogger(logger)
}

func Logger() *log.Logger {
	return logger
}

type LogStringer interface {
	LogString() string
}

type Validator = conv.Validator

func Validate(i interface{}) error {
	return conv.Validate(i)
}

func Next(ctx context.Context, req *Request) Responder {
	i, _ := ctx.Value(ctxutil.KeyNextHandler).(Handler)
	if i == nil {
		return nil
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ctxutil.KeyNextHandler, h)
}
