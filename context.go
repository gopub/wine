package wine

import (
	"context"
	"github.com/gopub/wine/ctxutil"
)

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
