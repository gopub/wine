package wine

import (
	"context"
	"fmt"
	"github.com/gopub/wine/internal/httpvfs"
	"net/http"
	"strings"

	"github.com/gopub/conv"
	"github.com/gopub/wine/router"
)

type metadata struct {
	Header *Header
}

func newMetadata() *metadata {
	return &metadata{
		Header: NewHeader(),
	}
}

func (m *metadata) clone() *metadata {
	return &metadata{
		Header: m.Header.Clone(),
	}
}

type Endpoint struct {
	*router.Endpoint
}

func (e *Endpoint) Header() *Header {
	return e.Metadata().(*metadata).Header
}

// Router implements routing function
type Router struct {
	*router.Router
	authChecker Handler
	md          *metadata
}

// NewRouter new a Router
func NewRouter() *Router {
	r := &Router{
		Router:      router.New(),
		authChecker: HandlerFunc(checkAuth),
		md:          newMetadata(),
	}
	r.bindSysHandlers()
	return r
}

func (r *Router) bindSysHandlers() {
	r.Get(endpointPath, r.listEndpoints)
	r.Get(datePath, handleDate)
	r.Bind(http.MethodGet, versionPath, HandleResponder(Text(http.StatusOK, "v1.26.5")))
	r.Get(uptimePath, newUptimeHandler())
	r.Handle(echoPath, handleEcho)
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

func (r *Router) Group(name string) *Router {
	nr := r.Router.Group(name)
	return &Router{
		Router:      nr,
		authChecker: r.authChecker,
		md:          r.md.clone(),
	}
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) UseHandlers(handlers ...Handler) *Router {
	nr := r.Router.Use(conv.ToList(handlers))
	return &Router{
		Router:      nr,
		authChecker: r.authChecker,
		md:          r.md.clone(),
	}
}

// Use is similar with UseHandlers
func (r *Router) Use(funcs ...HandlerFunc) *Router {
	nr := r.Router.Use(conv.ToList(funcs))
	return &Router{
		Router:      nr,
		authChecker: r.authChecker,
		md:          r.md.clone(),
	}
}

// bind binds method, path with handlers
func (r *Router) Bind(method, path string, handlers ...Handler) *Endpoint {
	return r.toEndpoint(r.Router.Bind(method, path, conv.ToList(handlers)))
}

// StaticFile binds path to a file
func (r *Router) StaticFile(path, filePath string) {
	r.Get(path, func(ctx context.Context, req *Request) Responder {
		return StaticFile(req.request, filePath)
	})
}

// StaticFileData binds path to data
func (r *Router) StaticFileData(path string, data []byte) {
	r.Get(path, func(ctx context.Context, req *Request) Responder {
		// http.FileServer can handle content range
		// sometimes it's not allowed to write all data to client in one time. e.g. play video
		return Handle(req.request, http.FileServer(httpvfs.NewFileSystem(data)))
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
func (r *Router) Handle(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind("", path, conv.ToList(funcs)))
}

// Get binds funcs to path with GET method
func (r *Router) Get(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodGet, path, conv.ToList(funcs)))
}

// Post binds funcs to path with POST method
func (r *Router) Post(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodPost, path, conv.ToList(funcs)))
}

// Put binds funcs to path with PUT method
func (r *Router) Put(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodPut, path, conv.ToList(funcs)))
}

// Patch binds funcs to path with PATCH method
func (r *Router) Patch(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodPatch, path, conv.ToList(funcs)))
}

// Delete binds funcs to path with DELETE method
func (r *Router) Delete(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodDelete, path, conv.ToList(funcs)))
}

// Options binds funcs to path with OPTIONS method
func (r *Router) Options(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodOptions, path, conv.ToList(funcs)))
}

// Head binds funcs to path with HEAD method
func (r *Router) Head(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodHead, path, conv.ToList(funcs)))
}

// Trace binds funcs to path with TRACE method
func (r *Router) Trace(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodTrace, path, conv.ToList(funcs)))
}

// Connect binds funcs to path with CONNECT method
func (r *Router) Connect(path string, funcs ...HandlerFunc) *Endpoint {
	return r.toEndpoint(r.Router.Bind(http.MethodConnect, path, conv.ToList(funcs)))
}

func (r *Router) listEndpoints(ctx context.Context, req *Request) Responder {
	var l []*router.Endpoint
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

func (r *Router) Header() *Header {
	return r.md.Header
}

func (r *Router) toEndpoint(e *router.Endpoint) *Endpoint {
	if e == nil {
		return nil
	}

	new := r.md.clone()
	if md, ok := e.Metadata().(*metadata); ok && md.Header != nil {
		for k, v := range md.Header.Header {
			new.Header.Header[k] = v
		}
	}
	e.SetMetadata(new)
	return &Endpoint{
		Endpoint: e,
	}

}
