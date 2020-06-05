package ws

import (
	"context"

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

func (r *Router) BindHandlers(path string, handlers ...Handler) *router.Route {
	return r.Router.Bind("", path, conv.ToList(handlers))
}

func (r *Router) Bind(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind("", path, conv.ToList(funcs))
}

func handleAuth(ctx context.Context, data interface{}) (interface{}, error) {
	if wine.GetUserID(ctx) <= 0 {
		return nil, errors.Unauthorized("")
	}
	return Next(ctx, data)
}
