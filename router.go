package wine

import (
	"context"
	"github.com/gopub/log"
	"net/http"
	"reflect"
	"strings"
)

// Router implements routing function
type Router struct {
	methodTrees map[string]*node
	basePath    string
	handlers    []Handler
}

// NewRouter new a Router
func NewRouter() *Router {
	r := &Router{}
	r.methodTrees = make(map[string]*node, 4)
	return r
}

// Group returns a new router whose basePath is r.basePath+path
func (r *Router) Group(path string) *Router {
	if path == "/" {
		logger.Panic("don't create default group \"/\"")
	}

	nr := &Router{}
	nr.methodTrees = r.methodTrees

	// support empty path
	if len(path) > 0 {
		nr.basePath = normalizePath(r.basePath + "/" + path)
	}

	nr.handlers = make([]Handler, len(r.handlers))
	copy(nr.handlers, r.handlers)
	return nr
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) UseHandlers(handlers ...Handler) *Router {
	if len(handlers) == 0 {
		return r
	}

	nr := &Router{
		methodTrees: r.methodTrees,
		basePath:    r.basePath,
	}
	nr.handlers = make([]Handler, len(r.handlers))
	copy(nr.handlers, r.handlers)

	for _, h := range handlers {
		found := false
		for _, rh := range nr.handlers {
			if reflect.TypeOf(rh).Comparable() && reflect.TypeOf(h).Comparable() && rh == h {
				found = true
				break
			}
		}

		if !found {
			nr.handlers = append(nr.handlers, h)
		}
	}

	return nr
}

// Use is similar with UseHandlers
func (r *Router) Use(funcs ...HandlerFunc) *Router {
	if len(funcs) == 0 {
		return r
	}
	return r.UseHandlers(convertToHandlers(funcs...)...)
}

// Match finds handlers and parses path parameters according to method and path
func (r *Router) Match(method string, path string) (*handlerList, map[string]string) {
	n := r.methodTrees[method]
	if n == nil {
		return nil, nil
	}

	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}
	return n.matchPath(segments)
}

// Bind binds method, path with handlers
func (r *Router) Bind(method string, path string, handlers ...Handler) {
	if path == "" {
		logger.Panic("invalid path")
	}

	if len(method) == 0 {
		logger.Panic("invalid method")
	}

	if len(handlers) == 0 {
		logger.Panic("no handler")
	}

	method = strings.ToUpper(method)
	root := r.methodTrees[method]
	if root == nil {
		root = &node{}
		root.t = _StaticNode
		r.methodTrees[method] = root
	}

	hs := make([]Handler, len(r.handlers))
	copy(hs, r.handlers)
	hs = append(hs, handlers...)
	hl := newHandlerList(hs)

	path = normalizePath(r.basePath + "/" + path)
	if path == "" {
		if root.handlers.Empty() {
			root.handlers = hl
		} else {
			panic("binding conflict: " + path)
		}
	} else {
		nodes := newNodeList(path, hl)
		if !root.add(nodes) {
			panic("binding conflict: " + path)
		}
	}
}

// Unbind unbinds method, path
func (r *Router) Unbind(method string, path string) {
	if path == "" {
		logger.Panic("invalid path")
	}

	if len(method) == 0 {
		logger.Panic("invalid method")
	}

	method = strings.ToUpper(method)
	root := r.methodTrees[method]
	if root == nil {
		root = &node{}
		root.t = _StaticNode
		r.methodTrees[method] = root
	}

	path = normalizePath(r.basePath + "/" + path)
	if path == "" {
		root.handlers = nil
		return
	}

	nodes := newNodeList(path, nil)
	if len(nodes) == 0 {
		logger.Panic("node list is empty")
	}

	if nodes[0].path != "" {
		emptyNode := &node{}
		emptyNode.t = _StaticNode
		nodes = append([]*node{emptyNode}, nodes...)
	}

	n := root.matchNodes(nodes)
	if n != nil {
		n.handlers = nil
		log.With("method", method, "path", path).Info("succeeded")
	}
	return
}

// StaticFile binds path to a file
func (r *Router) StaticFile(path, filePath string) {
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responsible {
		return File(req.HTTPRequest, filePath)
	})
	return
}

// StaticDir binds path to a directory
func (r *Router) StaticDir(path, dirPath string) {
	r.StaticFS(path, http.Dir(dirPath))
	return
}

// StaticFS binds path to an abstract file system
func (r *Router) StaticFS(path string, fs http.FileSystem) {
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
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responsible {
		return Handle(req.HTTPRequest, fileServer)
	})
	return
}

// Get binds funcs to path with GET method
func (r *Router) Get(path string, funcs ...HandlerFunc) {
	r.Bind("GET", path, convertToHandlers(funcs...)...)
}

// Post binds funcs to path with POST method
func (r *Router) Post(path string, funcs ...HandlerFunc) {
	r.Bind("POST", path, convertToHandlers(funcs...)...)
}

// Put binds funcs to path with PUT method
func (r *Router) Put(path string, funcs ...HandlerFunc) {
	r.Bind("PUT", path, convertToHandlers(funcs...)...)
}

// Delete binds funcs to path with DELETE method
func (r *Router) Delete(path string, funcs ...HandlerFunc) {
	r.Bind("DELETE", path, convertToHandlers(funcs...)...)
}

// Any binds funcs to path with GET/POST methods
func (r *Router) Any(path string, funcs ...HandlerFunc) {
	handlers := convertToHandlers(funcs...)
	r.Bind("GET", path, handlers...)
	r.Bind("POST", path, handlers...)
}

// BindControllers binds controllers
func (r *Router) BindControllers(controllers ...Controller) {
	for _, c := range controllers {
		for s, h := range c.RouteMap() {
			strs := strings.Fields(s)
			if len(strs) != 2 {
				logger.Panic("invalid route key: " + s)
			}
			methods := strings.Split(strs[0], "|")
			path := c.RoutePath() + "/" + strs[1]
			for _, method := range methods {
				r.Bind(method, path, h)
			}
		}
	}
}

// Print prints all path trees
func (r *Router) Print() {
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
