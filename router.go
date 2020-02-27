package wine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"

	"github.com/gopub/log"
	pathutil "github.com/gopub/wine/internal/path"
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
	r.Get(endpointPath, r.listEndpoints)
	return r
}

type endpointInfo struct {
	Method       string
	Path         string
	HandlerNames string
}

type endpointInfoList []*endpointInfo

func (l endpointInfoList) Len() int {
	return len(l)
}

func (l endpointInfoList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l endpointInfoList) Less(i, j int) bool {
	return l[i].Path < l[j].Path
}

func (r *Router) listEndpoints(ctx context.Context, req *Request, next Invoker) Responsible {
	endpoints := make(endpointInfoList, 0, 10)
	maxLenOfPath := 0
	for method, node := range r.methodTrees {
		pl := node.Endpoints(method)
		for _, p := range pl {
			endpoints = append(endpoints, p)
			if n := len(p.Path); n > maxLenOfPath {
				maxLenOfPath = n
			}
		}
	}
	sort.Sort(endpoints)

	b := new(strings.Builder)
	for i, e := range endpoints {
		format := fmt.Sprintf("%%3d. %%6s /%%-%ds %%s\n", maxLenOfPath)
		line := fmt.Sprintf(format, i+1, e.Method, e.Path, e.HandlerNames)
		b.WriteString(line)
	}
	return Text(http.StatusOK, b.String())
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
		nr.basePath = pathutil.Normalize(r.basePath + "/" + path)
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

// match finds handlers and parses path parameters according to method and path
func (r *Router) match(method string, path string) (*handlerList, map[string]string) {
	n := r.methodTrees[method]
	if n == nil {
		return nil, nil
	}

	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}
	hl, params := n.matchPath(segments)
	unescapedParams := make(map[string]string, len(params))
	for k, v := range params {
		uv, err := url.PathUnescape(v)
		if err != nil {
			logger.Errorf("Unescape path param %s: %v", v, err)
			unescapedParams[k] = v
		} else {
			unescapedParams[k] = uv
		}
	}
	return hl, unescapedParams
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
		root.t = staticNode
		r.methodTrees[method] = root
	}

	hs := make([]Handler, len(r.handlers))
	copy(hs, r.handlers)
	hs = append(hs, handlers...)
	hl := newHandlerList(hs)

	path = pathutil.Normalize(r.basePath + "/" + path)
	if path == "" {
		if root.handlers.Empty() {
			root.handlers = hl
		} else {
			root.Print(method, r.basePath)
			panic("binding conflict: " + path)
		}
	} else {
		nodes := newNodeList(path, hl)
		if !root.add(nodes) {
			root.Print(method, r.basePath)
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
		root.t = staticNode
		r.methodTrees[method] = root
	}

	path = pathutil.Normalize(r.basePath + "/" + path)
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
		emptyNode.t = staticNode
		nodes = append([]*node{emptyNode}, nodes...)
	}

	n := root.matchNodes(nodes)
	if n != nil {
		n.handlers = nil
		log.With("method", method, "path", path).Info("succeeded")
	}
}

// StaticFile binds path to a file
func (r *Router) StaticFile(path, filePath string) {
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responsible {
		return StaticFile(req.request, filePath)
	})
}

// StaticDir binds path to a directory
func (r *Router) StaticDir(path, dirPath string) {
	r.StaticFS(path, http.Dir(dirPath))
}

// StaticFS binds path to an abstract file system
func (r *Router) StaticFS(path string, fs http.FileSystem) {
	prefix := pathutil.Normalize(r.basePath + "/" + path)
	if len(prefix) == 0 {
		prefix = "/"
	} else if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	i := strings.Index(prefix, "*")
	if i > 0 {
		prefix = prefix[:i]
	} else {
		path = pathutil.Normalize(path + "/*")
	}

	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responsible {
		return Handle(req.request, fileServer)
	})
}

// Get binds funcs to path with GET method
func (r *Router) Get(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodGet, path, convertToHandlers(funcs...)...)
}

// Post binds funcs to path with POST method
func (r *Router) Post(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPost, path, convertToHandlers(funcs...)...)
}

// Put binds funcs to path with PUT method
func (r *Router) Put(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPut, path, convertToHandlers(funcs...)...)
}

// Patch binds funcs to path with PATCH method
func (r *Router) Patch(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPatch, path, convertToHandlers(funcs...)...)
}

// Delete binds funcs to path with DELETE method
func (r *Router) Delete(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodDelete, path, convertToHandlers(funcs...)...)
}

// Options binds funcs to path with OPTIONS method
func (r *Router) Options(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodOptions, path, convertToHandlers(funcs...)...)
}

// Head binds funcs to path with HEAD method
func (r *Router) Head(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodHead, path, convertToHandlers(funcs...)...)
}

// Trace binds funcs to path with TRACE method
func (r *Router) Trace(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodTrace, path, convertToHandlers(funcs...)...)
}

// Connect binds funcs to path with CONNECT method
func (r *Router) Connect(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodConnect, path, convertToHandlers(funcs...)...)
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
