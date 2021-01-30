package ctxutil

import (
	"context"
	"net/http"

	"github.com/gopub/log"
	"github.com/gopub/wine/httpvalue"
	"github.com/gopub/wine/internal/template"
)

type Key int

// Context keys
const (
	keyStart Key = iota

	KeyNextHandler
	KeyTemplateManager
	KeyUserID
	KeyTraceID
	KeyUser
	KeySudo
	KeyRequestHeader

	keyEnd
)

func GetRequestHeader(ctx context.Context) http.Header {
	h, _ := ctx.Value(KeyRequestHeader).(http.Header)
	return h
}

func WithRequestHeader(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, KeyRequestHeader, h)
}

func GetTraceID(ctx context.Context) string {
	id, ok := ctx.Value(KeyTraceID).(string)
	if ok {
		return id
	}
	return GetRequestHeader(ctx).Get(httpvalue.CustomTraceID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, KeyTraceID, traceID)
}

func GetTemplateManager(ctx context.Context) *template.Manager {
	v, _ := ctx.Value(KeyTemplateManager).(*template.Manager)
	return v
}

func WithTemplateManager(ctx context.Context, m *template.Manager) context.Context {
	return context.WithValue(ctx, KeyTemplateManager, m)
}

func Detach(ctx context.Context) context.Context {
	newCtx := context.Background()
	if l := log.FromContext(ctx); l != nil {
		newCtx = log.BuildContext(newCtx, l)
	}
	for k := keyStart; k < keyEnd; k++ {
		if v := ctx.Value(k); v != nil {
			newCtx = context.WithValue(newCtx, k, v)
		}
	}
	return newCtx
}

func GetUserID(ctx context.Context) int64 {
	id, _ := ctx.Value(KeyUserID).(int64)
	return id
}

func WithUserID(ctx context.Context, id int64) context.Context {
	if id == 0 {
		return ctx
	}
	return context.WithValue(ctx, KeyUserID, id)
}

func GetUser(ctx context.Context) interface{} {
	return ctx.Value(KeyUser)
}

func WithUser(ctx context.Context, u interface{}) context.Context {
	return context.WithValue(ctx, KeyUser, u)
}

func WithSudo(ctx context.Context) context.Context {
	return context.WithValue(ctx, KeySudo, true)
}

func IsSudo(ctx context.Context) bool {
	b, _ := ctx.Value(KeySudo).(bool)
	return b
}

func GetAppID(ctx context.Context) string {
	return GetRequestHeader(ctx).Get(httpvalue.CustomAppID)
}

func GetDeviceID(ctx context.Context) string {
	return GetRequestHeader(ctx).Get(httpvalue.CustomDeviceID)
}
