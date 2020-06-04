package ws

import (
	"github.com/gopub/wine/router"
)

// Router implements routing function
type Router struct {
	*router.Router
	authHandler Handler
}

// NewRouter new a Router
func NewRouter() *Router {
	r := &Router{
		Router: router.New(),
	}
	r.bindSysHandlers()
	return r
}

func (r *Router) bindSysHandlers() {
	//r.Get(endpointPath, r.listEndpoints)
	//r.Get(datePath, handleDate)
	//r.Bind(http.MethodGet, versionPath, HandleResponder(Text(http.StatusOK, "v1.23.0.2")))
	//r.Get(uptimePath, newUptimeHandler())
	//r.Handle(echoPath, handleEcho)
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

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) UseHandlers(handlers ...Handler) *Router {
	return &Router{
		Router:      r.Router.Use(linkHandlers(handlers)),
		authHandler: r.authHandler,
	}
}

// Use is similar with UseHandlers
func (r *Router) Use(funcs ...HandlerFunc) *Router {
	return &Router{
		Router:      r.Router.Use(linkHandlerFuncs(funcs)),
		authHandler: r.authHandler,
	}
}

// Bind binds method, path with handlers
func (r *Router) BindHandlers(path string, handlers ...Handler) *router.Route {
	return r.Router.Bind("", path, linkHandlers(handlers))
}

// Handle binds funcs to path with any(wildcard) method
func (r *Router) Bind(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind("", path, linkHandlerFuncs(funcs))
}
