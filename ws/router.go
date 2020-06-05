package ws

import (
	"container/list"
	"context"
	"reflect"
	"runtime"

	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	"github.com/gopub/wine/router"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req interface{}) (interface{}, error)
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req interface{}) (interface{}, error)

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return h(ctx, req)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

type handlerElem list.Element

func (h *handlerElem) Next() *handlerElem {
	return (*handlerElem)((*list.Element)(h).Next())
}

func (h *handlerElem) HandleRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return h.Value.(Handler).HandleRequest(withNextHandler(ctx, h.Next()), req)
}

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
