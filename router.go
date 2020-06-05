package wine

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gopub/conv"
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
		Router:      router.New(),
		authHandler: HandlerFunc(handleAuth),
	}
	r.bindSysHandlers()
	return r
}

func (r *Router) bindSysHandlers() {
	r.Get(endpointPath, r.listEndpoints)
	r.Get(datePath, handleDate)
	r.Bind(http.MethodGet, versionPath, HandleResponder(Text(http.StatusOK, "v1.23.0.2")))
	r.Get(uptimePath, newUptimeHandler())
	r.Handle(echoPath, handleEcho)
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

func (r *Router) Group(name string) *Router {
	nr := r.Router.Group(name)
	return &Router{
		Router:      nr,
		authHandler: r.authHandler,
	}
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) UseHandlers(handlers ...Handler) *Router {
	nr := r.Router.Use(conv.ToList(handlers))
	return &Router{
		Router:      nr,
		authHandler: r.authHandler,
	}
}

// Use is similar with UseHandlers
func (r *Router) Use(funcs ...HandlerFunc) *Router {
	nr := r.Router.Use(conv.ToList(funcs))
	return &Router{
		Router:      nr,
		authHandler: r.authHandler,
	}
}

// bind binds method, path with handlers
func (r *Router) Bind(method, path string, handlers ...Handler) *router.Route {
	return r.Router.Bind(method, path, conv.ToList(handlers))
}

// StaticFile binds path to a file
func (r *Router) StaticFile(path, filePath string) {
	r.Get(path, func(ctx context.Context, req *Request) Responder {
		return StaticFile(req.request, filePath)
	})
}

// StaticDir binds path to a directory
func (r *Router) StaticDir(path, dirPath string) {
	r.StaticFS(path, http.Dir(dirPath))
}

// StaticFS binds path to an abstract file system
func (r *Router) StaticFS(path string, fs http.FileSystem) {
	prefix := router.Normalize(r.BasePath() + "/" + path)
	if prefix == "" {
		prefix = "/"
	} else if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	i := strings.Index(prefix, "*")
	if i > 0 {
		prefix = prefix[:i]
	} else {
		path = router.Normalize(path + "/*")
	}

	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	r.Get(path, func(ctx context.Context, req *Request) Responder {
		return Handle(req.request, fileServer)
	})
}

// Handle binds funcs to path with any(wildcard) method
func (r *Router) Handle(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind("", path, conv.ToList(funcs))
}

// Get binds funcs to path with GET method
func (r *Router) Get(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodGet, path, conv.ToList(funcs))
}

// Post binds funcs to path with POST method
func (r *Router) Post(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodPost, path, conv.ToList(funcs))
}

// Put binds funcs to path with PUT method
func (r *Router) Put(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodPut, path, conv.ToList(funcs))
}

// Patch binds funcs to path with PATCH method
func (r *Router) Patch(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodPatch, path, conv.ToList(funcs))
}

// Delete binds funcs to path with DELETE method
func (r *Router) Delete(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodDelete, path, conv.ToList(funcs))
}

// Options binds funcs to path with OPTIONS method
func (r *Router) Options(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodOptions, path, conv.ToList(funcs))
}

// Head binds funcs to path with HEAD method
func (r *Router) Head(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodHead, path, conv.ToList(funcs))
}

// Trace binds funcs to path with TRACE method
func (r *Router) Trace(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodTrace, path, conv.ToList(funcs))
}

// Connect binds funcs to path with CONNECT method
func (r *Router) Connect(path string, funcs ...HandlerFunc) *router.Route {
	return r.Router.Bind(http.MethodConnect, path, conv.ToList(funcs))
}

func (r *Router) listEndpoints(ctx context.Context, req *Request) Responder {
	var l []*router.Route
	maxLenOfPath := 0
	all := req.params.Bool("all")
	for _, node := range r.ListRoutes() {
		if !all && reservedPaths[node.Path()] {
			continue
		}
		l = append(l, node)
		if n := len(node.Path()); n > maxLenOfPath {
			maxLenOfPath = n
		}
	}
	b := new(strings.Builder)
	for i, n := range l {
		format := fmt.Sprintf("%%3d. %%6s /%%-%ds %%s", maxLenOfPath)
		line := fmt.Sprintf(format, i+1, n.Scope, n.Path(), n.HandlerPath())
		b.WriteString(line)
		if n.Description() != "" {
			b.WriteString(" #")
			b.WriteString(n.Description())
		}
		b.WriteString("\n")
	}
	return Text(http.StatusOK, b.String())
}
