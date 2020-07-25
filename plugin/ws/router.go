package ws

import (
	"context"
	"time"

	"github.com/gopub/types"

	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
)

type Router struct {
	*router.Router
	authHandler Handler
}

func NewRouter() *Router {
	r := &Router{
		Router:      router.New(),
		authHandler: HandlerFunc(handleAuth),
	}
	r.Bind("ws.getDate", handleDate)
	return r
}

func (r *Router) SetAuthHandler(h Handler) {
	r.authHandler = h
}

func (r *Router) Auth() *Router {
	if r.ContainsHandler(r.authHandler) {
		return r
	}
	return r.UseHandlers(r.authHandler)
}

func (r *Router) UseHandlers(handlers ...Handler) *Router {
	return &Router{
		Router:      r.Router.Use(conv.ToList(handlers)),
		authHandler: r.authHandler,
	}
}

func (r *Router) Use(funcs ...HandlerFunc) *Router {
	return &Router{
		Router:      r.Router.Use(conv.ToList(funcs)),
		authHandler: r.authHandler,
	}
}

func (r *Router) BindHandlers(path string, handlers ...Handler) *router.Endpoint {
	return r.Router.Bind("", path, conv.ToList(handlers))
}

func (r *Router) Bind(path string, funcs ...HandlerFunc) *router.Endpoint {
	return r.Router.Bind("", path, conv.ToList(funcs))
}

func handleAuth(ctx context.Context, params interface{}) (interface{}, error) {
	if wine.GetUserID(ctx) <= 0 {
		return nil, errors.Unauthorized("")
	}
	return Next(ctx, params)
}

func handleDate(_ context.Context, _ interface{}) (interface{}, error) {
	t := time.Now()
	res := types.M{
		"timestamp": t.Unix(),
		"time":      t.Format("15:04:05"),
		"date":      t.Format("2006-01-02"),
		"zone":      t.Format("-0700"),
		"weekday":   t.Format("Mon"),
		"month":     t.Format("Jan"),
	}
	return res, nil
}

func Next(ctx context.Context, req interface{}) (interface{}, error) {
	i, _ := ctx.Value(ckNextHandler).(Handler)
	if i == nil {
		return nil, errors.NotImplemented("")
	}
	return i.HandleRequest(ctx, req)
}

func withNextHandler(ctx context.Context, h Handler) context.Context {
	return context.WithValue(ctx, ckNextHandler, h)
}
