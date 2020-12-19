package wine

import (
	"context"
	"net/http"

	"github.com/gopub/log"
	"github.com/gopub/wine/internal/template"
)

type contextKey int

// Context keys
const (
	ckNextHandler contextKey = iota + 1
	ckTemplateManager
	ckUserID
	ckTraceID
	ckUser
	ckSudo
	ckRequestHeader
)

func Next(ctx context.Context, req *Request) Responder {
	i, _ := ctx.Value(ckNextHandler).(Handler)
	if i == nil {
		return nil
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ckNextHandler, h)
}

func getTemplateManager(ctx context.Context) *template.Manager {
	v, _ := ctx.Value(ckTemplateManager).(*template.Manager)
	return v
}

func withTemplateManager(ctx context.Context, m *template.Manager) context.Context {
	return context.WithValue(ctx, ckTemplateManager, m)
}

func GetUserID(ctx context.Context) int64 {
	id, _ := ctx.Value(ckUserID).(int64)
	return id
}

func WithUserID(ctx context.Context, id int64) context.Context {
	if id == 0 {
		return ctx
	}
	return context.WithValue(ctx, ckUserID, id)
}

func GetUser(ctx context.Context) interface{} {
	return ctx.Value(ckUser)
}

func WithUser(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, ckUser, u)
}

func GetRequestHeader(ctx context.Context) http.Header {
	h, _ := ctx.Value(ckRequestHeader).(http.Header)
	return h
}

func withRequestHeader(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, ckRequestHeader, h)
}

func GetTraceID(ctx context.Context) string {
	id, _ := ctx.Value(ckTraceID).(string)
	return id
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, ckTraceID, traceID)
}

func WithSudo(ctx context.Context) context.Context {
	return context.WithValue(ctx, ckSudo, true)
}

func IsSudo(ctx context.Context) bool {
	b, _ := ctx.Value(ckSudo).(bool)
	return b
}

func DetachContext(ctx context.Context) context.Context {
	newCtx := context.Background()
	if l := log.FromContext(ctx); l != nil {
		newCtx = log.BuildContext(newCtx, l)
	}
	if m := getTemplateManager(ctx); m != nil {
		newCtx = withTemplateManager(ctx, m)
	}
	if h := GetRequestHeader(ctx); h != nil {
		newCtx = withRequestHeader(newCtx, h)
	}
	if traceID := GetTraceID(ctx); traceID != "" {
		newCtx = WithTraceID(newCtx, traceID)
	}
	if uid := GetUserID(ctx); uid > 0 {
		newCtx = WithUserID(newCtx, uid)
	}
	if u := GetUser(ctx); u != nil {
		newCtx = WithUser(newCtx, u)
	}
	if IsSudo(ctx) {
		newCtx = WithSudo(newCtx)
	}
	return newCtx
}
