package wine

import (
	"context"
	"net/http"

	contextpkg "github.com/gopub/wine/internal/context"
)

func Next(ctx context.Context, req *Request) Responder {
	i, _ := ctx.Value(contextpkg.KeyNextHandler).(Handler)
	if i == nil {
		return nil
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, contextpkg.KeyNextHandler, h)
}

func GetUserID(ctx context.Context) int64 {
	id, _ := ctx.Value(contextpkg.KeyUserID).(int64)
	return id
}

func WithUserID(ctx context.Context, id int64) context.Context {
	if id == 0 {
		return ctx
	}
	return context.WithValue(ctx, contextpkg.KeyUserID, id)
}

func GetUser(ctx context.Context) interface{} {
	return ctx.Value(contextpkg.KeyUser)
}

func WithUser(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, contextpkg.KeyUser, u)
}

func GetRequestHeader(ctx context.Context) http.Header {
	return contextpkg.GetRequestHeader(ctx)
}

func GetTraceID(ctx context.Context) string {
	return contextpkg.GetTraceID(ctx)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return contextpkg.WithTraceID(ctx, traceID)
}

func WithSudo(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextpkg.KeySudo, true)
}

func IsSudo(ctx context.Context) bool {
	b, _ := ctx.Value(contextpkg.KeySudo).(bool)
	return b
}

func GetAppID(ctx context.Context) string {
	return contextpkg.GetTraceID(ctx)
}

func WithAppID(ctx context.Context, appID string) context.Context {
	return contextpkg.WithAppID(ctx, appID)
}

func GetDeviceID(ctx context.Context) string {
	return contextpkg.GetDeviceID(ctx)
}

func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return contextpkg.WithDeviceID(ctx, deviceID)
}

func DetachContext(ctx context.Context) context.Context {
	return contextpkg.Detach(ctx)
}
