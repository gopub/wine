package wine

import (
	"net/http"
	"strings"
)

// Router routes path to handlers
type Router interface {
	Group(path string) Router
	UseHandlers(handlers ...Handler) Router
	Use(funcs ...HandlerFunc) Router
	StaticFile(path, filePath string)
	StaticDir(path, filePath string)
	StaticFS(path string, fs http.FileSystem)
	Bind(method, path string, handlers ...Handler)
	Match(method, path string) (handlers []Handler, params map[string]string)

	HandleGet(path string, handlers ...Handler)
	HandlePost(path string, handlers ...Handler)
	HandleDelete(path string, handlers ...Handler)
	HandlePut(path string, handlers ...Handler)
	HandleAny(path string, handlers ...Handler)

	Get(path string, handlers ...HandlerFunc)
	Post(path string, handlers ...HandlerFunc)
	Put(path string, handlers ...HandlerFunc)
	Delete(path string, handlers ...HandlerFunc)
	Any(path string, handlers ...HandlerFunc)
}

// DefaultRouter is a default implementation of Router interface
type DefaultRouter struct {
	methodTrees map[string]*node
	basePath    string
	handlers    []Handler
}

// NewDefaultRouter new a DefaultRouter
func NewDefaultRouter() *DefaultRouter {
	r := &DefaultRouter{}
	r.methodTrees = make(map[string]*node, 4)
	return r
}

// Group returns a new router whose basePath is r.basePath+path
func (r *DefaultRouter) Group(path string) Router {
	if len(path) == 0 {
		panic("[WINE] group path is empty")
	}

	if path == "/" {
		panic("[WINE] don't create default group \"/\"")
	}

	nr := &DefaultRouter{}
	nr.methodTrees = r.methodTrees
	nr.basePath = normalizePath(r.basePath + "/" + path)
	nr.handlers = make([]Handler, len(r.handlers))
	copy(nr.handlers, r.handlers)
	return nr
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *DefaultRouter) UseHandlers(handlers ...Handler) Router {
	if len(handlers) == 0 {
		return r
	}
	r.handlers = append(r.handlers, handlers...)
	return r
}

// Use is similar with UseHandlers
func (r *DefaultRouter) Use(funcs ...HandlerFunc) Router {
	if len(funcs) == 0 {
		return r
	}
	return r.UseHandlers(convertToHandlers(funcs...)...)
}

// Match finds handlers and parses path parameters according to method and path
func (r *DefaultRouter) Match(method string, path string) (handlers []Handler, params map[string]string) {
	n := r.methodTrees[method]
	if n == nil {
		return
	}

	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}
	return n.match(segments)
}

// Bind binds method, path with handlers
func (r *DefaultRouter) Bind(method string, path string, handlers ...Handler) {
	if path == "" {
		panic("[WINE] invalid path")
	}

	if len(method) == 0 {
		panic("[WINE] invalid method")
	}

	if len(handlers) == 0 {
		panic("[WINE] no handler")
	}

	n := r.methodTrees[method]
	if n == nil {
		n = &node{}
		n.t = staticNode
		r.methodTrees[method] = n
	}

	hs := make([]Handler, len(r.handlers))
	copy(hs, r.handlers)
	hs = append(hs, handlers...)

	path = normalizePath(r.basePath + "/" + path)
	if path == "" {
		if len(n.handlers) == 0 {
			n.handlers = hs
		} else {
			panic("binding conflict: " + path)
		}
	} else {
		nodes := newNodeList(path, hs...)
		if !n.add(nodes) {
			panic("binding conflict: " + path)
		}
	}
}

// StaticFile binds path to a file
func (r *DefaultRouter) StaticFile(path, filePath string) {
	r.Get(path, func(c Context) {
		c.File(filePath)
	})
	return
}

// StaticDir binds path to a directory
func (r *DefaultRouter) StaticDir(path, dirPath string) {
	r.StaticFS(path, http.Dir(dirPath))
	return
}

// StaticFS binds path to an abstract file system
func (r *DefaultRouter) StaticFS(path string, fs http.FileSystem) {
	prefix := normalizePath(r.basePath + "/" + path)
	if len(prefix) == 0 {
		prefix = "/"
	} else if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	i := strings.Index(prefix, "*")
	if i > 0 {
		prefix = prefix[:i]
	} else {
		path = normalizePath(path + "/*")
	}

	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	r.Get(path, func(c Context) {
		c.ServeHTTP(fileServer)
	})
	return
}

// HandleGet binds handlers to path with GET method
func (r *DefaultRouter) HandleGet(path string, handlers ...Handler) {
	r.Bind("GET", path, handlers...)
	return
}

// HandlePost binds handlers to path with POST method
func (r *DefaultRouter) HandlePost(path string, handlers ...Handler) {
	r.Bind("POST", path, handlers...)
	return
}

// HandlePut binds handlers to path with PUT method
func (r *DefaultRouter) HandlePut(path string, handlers ...Handler) {
	r.Bind("PUT", path, handlers...)
	return
}

// HandleDelete binds handlers to path with DELETE method
func (r *DefaultRouter) HandleDelete(path string, handlers ...Handler) {
	r.Bind("DELETE", path, handlers...)
	return
}

// HandleAny binds handlers to path with GET/POST/PUT methods
func (r *DefaultRouter) HandleAny(path string, handlers ...Handler) {
	r.HandleGet(path, handlers...)
	r.HandlePost(path, handlers...)
	r.HandlePut(path, handlers...)
	return
}

// Get binds funcs to path with GET method
func (r *DefaultRouter) Get(path string, funcs ...HandlerFunc) {
	r.HandleGet(path, convertToHandlers(funcs...)...)
}

// Post binds funcs to path with POST method
func (r *DefaultRouter) Post(path string, funcs ...HandlerFunc) {
	r.HandlePost(path, convertToHandlers(funcs...)...)
}

// Put binds funcs to path with PUT method
func (r *DefaultRouter) Put(path string, funcs ...HandlerFunc) {
	r.HandlePut(path, convertToHandlers(funcs...)...)
}

// Delete binds funcs to path with DELETE method
func (r *DefaultRouter) Delete(path string, funcs ...HandlerFunc) {
	r.HandleDelete(path, convertToHandlers(funcs...)...)
}

// Any binds funcs to path with GET/POST/PUT methods
func (r *DefaultRouter) Any(path string, funcs ...HandlerFunc) {
	r.HandleAny(path, convertToHandlers(funcs...)...)
}

// Print prints all path trees
func (r *DefaultRouter) Print() {
	for m, n := range r.methodTrees {
		n.Print(m, "/")
	}
}

func convertToHandlers(funcs ...HandlerFunc) []Handler {
	handlers := make([]Handler, len(funcs))
	for i, h := range funcs {
		handlers[i] = h
	}
	return handlers
}
