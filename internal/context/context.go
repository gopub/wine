package context

import (
	"context"
	"net/http"

	"github.com/gopub/log"
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
	KeyAppID
	KeyDeviceID

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
	id, _ := ctx.Value(KeyTraceID).(string)
	return id
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

func GetAppID(ctx context.Context) string {
	s, _ := ctx.Value(KeyAppID).(string)
	return s
}

func WithAppID(ctx context.Context, appID string) context.Context {
	return context.WithValue(ctx, KeyAppID, appID)
}

func GetDeviceID(ctx context.Context) string {
	s, _ := ctx.Value(KeyDeviceID).(string)
	return s
}

func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, KeyDeviceID, deviceID)
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
