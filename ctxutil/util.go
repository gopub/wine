package ctxutil

import (
	"context"
	"net/http"

	"github.com/gopub/wine/httpvalue"

	"github.com/gopub/log/v2"
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
	KeyBasicUser

	keyEnd
)

type User interface {
	GetID() int64
}

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
	if id > 0 {
		return id
	}

	if u := GetUser(ctx); u != nil {
		return u.GetID()
	}
	return 0
}

func WithUserID(ctx context.Context, id int64) context.Context {
	if id == 0 {
		return ctx
	}
	return context.WithValue(ctx, KeyUserID, id)
}

func GetUser(ctx context.Context) User {
	u, _ := ctx.Value(KeyUser).(User)
	return u
}

func WithBasicUser(ctx context.Context, u string) context.Context {
	return context.WithValue(ctx, KeyUser, u)
}

func GetBasicUser(ctx context.Context) string {
	u, _ := ctx.Value(KeyUser).(string)
	return u
}

func WithUser(ctx context.Context, u User) context.Context {
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
