package wine

import (
	"context"

	"github.com/gopub/conv"
	"github.com/gopub/log"
	"github.com/gopub/wine/ctxutil"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/router"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine")
	logger.SetFlags(log.LstdFlags - log.Lfunction - log.Lshortfile)
	router.SetLogger(logger)
	respond.SetLogger(logger)
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
