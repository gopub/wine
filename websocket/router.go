package websocket

import (
	"context"
	"time"

	"github.com/gopub/wine/ctxutil"
	"github.com/gopub/wine/router"

	"github.com/gopub/types"

	"github.com/gopub/conv"
	"github.com/gopub/errors"
)

type Router struct {
	*router.Router
	authChecker Handler
}

func NewRouter() *Router {
	r := &Router{
		Router:      router.New(),
		authChecker: HandlerFunc(checkAuth),
	}
	r.Bind("websocket.getDate", handleDate)
	return r
}

// SetAuthChecker check if the request is authenticated.
// AuthChecker should not do authenticating which is supposed to be done ahead.
// Authentication can be done in PreHandler, which can identify every incoming request.
// Some endpoints are public no matter authenticated or not, however some may need to check authentication.
// In PreHandler, authentication may succeed or fail, it doesn't matter.
// AuthChecker will fail all non-authenticated requests
// Regarding to authorization, it's related to business logic, so the router won't handle this.
func (r *Router) SetAuthChecker(h Handler) {
	r.authChecker = h
}

func (r *Router) RequireAuth() *Router {
	if r.ContainsHandler(r.authChecker) {
		return r
	}
	return r.UseHandlers(r.authChecker)
}

func (r *Router) UseHandlers(handlers ...Handler) *Router {
	return &Router{
		Router:      r.Router.Use(conv.ToList(handlers)),
		authChecker: r.authChecker,
	}
}

func (r *Router) Use(funcs ...HandlerFunc) *Router {
	return &Router{
		Router:      r.Router.Use(conv.ToList(funcs)),
		authChecker: r.authChecker,
	}
}

func (r *Router) BindHandlers(path string, handlers ...Handler) *router.Endpoint {
	return r.Router.Bind("", path, conv.ToList(handlers))
}

func (r *Router) Bind(path string, funcs ...HandlerFunc) *router.Endpoint {
	return r.Router.Bind("", path, conv.ToList(funcs))
}

func checkAuth(ctx context.Context, params interface{}) (interface{}, error) {
	if ctxutil.GetUserID(ctx) <= 0 {
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
