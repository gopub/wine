package wine

import (
	"context"

	"github.com/gopub/wine/internal/template"

	"github.com/gopub/log"
	"github.com/gopub/types"
)

type contextKey int

// Context keys
const (
	ckHandlerChain contextKey = iota + 1
	ckBasicAuthUser
	ckTemplateManager
	ckSessionID
	ckRemoteAddr
	ckCoordinate
	ckAccessToken
	ckUserID
	ckTraceID
	ckUser
	ckDeviceID
	ckSudo
)

func Next(ctx context.Context, req *Request) Responder {
	i, _ := ctx.Value(ckHandlerChain).(*handlerChain)
	if i == nil {
		return nil
	}
	return i.HandleRequest(ctx, req)
}

func withHandlerChain(ctx context.Context, h *handlerChain) context.Context {
	return context.WithValue(ctx, ckHandlerChain, h)
}

func GetBasicAuthUser(ctx context.Context) string {
	user, _ := ctx.Value(ckBasicAuthUser).(string)
	return user
}

func withBasicAuthUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, ckBasicAuthUser, user)
}

func GetSessionID(ctx context.Context) string {
	sid, _ := ctx.Value(ckSessionID).(string)
	return sid
}

func withSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, ckSessionID, sid)
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

func GetAccessToken(ctx context.Context) string {
	token, _ := ctx.Value(ckAccessToken).(string)
	return token
}

func WithAccessToken(ctx context.Context, token string) context.Context {
	if token == "" {
		return ctx
	}
	return context.WithValue(ctx, ckAccessToken, token)
}

func GetRemoteAddr(ctx context.Context) string {
	ip, _ := ctx.Value(ckRemoteAddr).(string)
	return ip
}

func WithRemoteAddr(ctx context.Context, addr string) context.Context {
	if addr == "" {
		return ctx
	}
	return context.WithValue(ctx, ckRemoteAddr, addr)
}

func GetDeviceID(ctx context.Context) string {
	id, _ := ctx.Value(ckDeviceID).(string)
	return id
}

func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	if deviceID == "" {
		return ctx
	}
	return context.WithValue(ctx, ckDeviceID, deviceID)
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

func GetCoordinate(ctx context.Context) *types.Point {
	p, _ := ctx.Value(ckCoordinate).(*types.Point)
	return p
}

func WithCoordinate(ctx context.Context, p *types.Point) context.Context {
	if p == nil {
		return ctx
	}
	return context.WithValue(ctx, ckCoordinate, p)
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
	if u := GetBasicAuthUser(ctx); u != "" {
		newCtx = withBasicAuthUser(ctx, u)
	}
	if sid := GetSessionID(ctx); sid != "" {
		newCtx = withSessionID(ctx, sid)
	}
	if token := GetAccessToken(ctx); token != "" {
		newCtx = WithAccessToken(newCtx, token)
	}
	if deviceID := GetDeviceID(ctx); deviceID != "" {
		newCtx = WithDeviceID(newCtx, deviceID)
	}
	if c := GetCoordinate(ctx); c != nil {
		newCtx = WithCoordinate(newCtx, c)
	}
	if addr := GetRemoteAddr(ctx); addr != "" {
		newCtx = WithRemoteAddr(newCtx, addr)
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
