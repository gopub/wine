package wine

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"

	pathpkg "github.com/gopub/wine/internal/path"
)

// Router implements routing function
type Router struct {
	root         *pathpkg.Node
	methodToRoot map[string]*pathpkg.Node
	basePath     string
	handlers     []Handler
	authHandler  Handler
}

// NewRouter new a Router
func NewRouter() *Router {
	r := &Router{
		root:         pathpkg.NewEmptyNode(),
		methodToRoot: make(map[string]*pathpkg.Node, 4),
	}
	r.bindSysHandlers()
	return r
}

func (r *Router) bindSysHandlers() {
	r.Get(endpointPath, r.listEndpoints)
	r.Get(sysDatePath, handleDate)
	r.Handle(echoPath, handleEcho)
}

func (r *Router) clone() *Router {
	nr := &Router{
		root:         r.root,
		methodToRoot: r.methodToRoot,
		basePath:     r.basePath,
		authHandler:  HandlerFunc(handleAuth),
	}
	nr.handlers = make([]Handler, len(r.handlers))
	copy(nr.handlers, r.handlers)
	return nr
}

func (r *Router) SetAuthHandler(h Handler) {
	r.authHandler = h
}

func (r *Router) Auth() *Router {
	for _, h := range r.handlers {
		if h == r.authHandler {
			return r
		}
	}
	return r.UseHandlers(r.authHandler)
}

// Group returns a new router whose basePath is r.basePath+path
func (r *Router) Group(path string) *Router {
	if path == "/" {
		logger.Panic(`Not allowed to create group "/"`)
	}

	nr := r.clone()
	// support empty path
	if len(path) > 0 {
		nr.basePath = pathpkg.Normalize(r.basePath + "/" + path)
	}
	return nr
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) UseHandlers(handlers ...Handler) *Router {
	nr := r.clone()
	for _, h := range handlers {
		if h == nil {
			log.Fatalf("Handler is nil")
		}
		found := false
		for _, rh := range nr.handlers {
			if equal(h, rh) {
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
	return r.UseHandlers(toHandlers(funcs...)...)
}

// match finds handlers and parses path parameters according to method and path
func (r *Router) match(method string, path string) (*list.List, map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}

	root := r.methodToRoot[method]
	if root == nil {
		root = r.root
	}

	m, params := root.Match(segments...)
	if m == nil && root != r.root {
		m, params = r.root.Match(segments...)
	}

	if m == nil {
		return nil, map[string]string{}
	}

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
	return m.Handlers(), unescapedParams
}

func (r *Router) matchMethods(path string) []string {
	var methods []string
	for m := range r.methodToRoot {
		if handlers, _ := r.match(m, path); handlers != nil && handlers.Len() > 0 {
			methods = append(methods, m)
		}
	}
	return methods
}

// Bind binds method, path with handlers
func (r *Router) Bind(method, path string, handlers ...Handler) {
	if path == "" {
		logger.Panic("Empty path")
	}

	if method == "" {
		logger.Panic("Empty method")
	}

	if len(handlers) == 0 {
		logger.Panic("No handlers")
	}

	method = strings.ToUpper(method)
	root := r.getRoot(method)
	hl := r.createHandlerList(handlers)
	path = pathpkg.Normalize(r.basePath + "/" + path)
	if path == "" {
		if r.root.IsEndpoint() {
			logger.Panicf("Conflict: %s", r.basePath)
		} else if root.IsEndpoint() {
			logger.Panicf("Conflict: %s, %s", method, r.basePath)
		} else {
			root.SetHandlers(hl)
		}
	} else {
		nodeList := pathpkg.NewNodeList(path, hl)
		if pair := r.root.Conflict(nodeList); pair != nil {
			first := pair.First.(*pathpkg.Node).Path()
			second := pair.Second.(*pathpkg.Node).Path()
			logger.Panicf("Conflict: %s, %s %s", first, method, second)
		}
		root.Add(nodeList)
	}
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
	prefix := pathpkg.Normalize(r.basePath + "/" + path)
	if prefix == "" {
		prefix = "/"
	} else if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	i := strings.Index(prefix, "*")
	if i > 0 {
		prefix = prefix[:i]
	} else {
		path = pathpkg.Normalize(path + "/*")
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
func (r *Router) Handle(path string, funcs ...HandlerFunc) {
	if path == "" {
		logger.Panic("Empty path")
	}
	handlers := toHandlers(funcs...)
	if len(handlers) == 0 {
		logger.Panic("No handlers")
	}

	hl := r.createHandlerList(handlers)
	path = pathpkg.Normalize(r.basePath + "/" + path)
	if path == "" {
		if r.root.IsEndpoint() {
			logger.Panicf("Conflict: %s", r.basePath)
		} else {
			r.root.SetHandlers(hl)
		}
	} else {
		nodeList := pathpkg.NewNodeList(path, hl)
		r.root.Add(nodeList)
	}
}

// Get binds funcs to path with GET method
func (r *Router) Get(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodGet, path, toHandlers(funcs...)...)
}

// Post binds funcs to path with POST method
func (r *Router) Post(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPost, path, toHandlers(funcs...)...)
}

// Put binds funcs to path with PUT method
func (r *Router) Put(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPut, path, toHandlers(funcs...)...)
}

// Patch binds funcs to path with PATCH method
func (r *Router) Patch(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodPatch, path, toHandlers(funcs...)...)
}

// Delete binds funcs to path with DELETE method
func (r *Router) Delete(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodDelete, path, toHandlers(funcs...)...)
}

// Options binds funcs to path with OPTIONS method
func (r *Router) Options(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodOptions, path, toHandlers(funcs...)...)
}

// Head binds funcs to path with HEAD method
func (r *Router) Head(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodHead, path, toHandlers(funcs...)...)
}

// Trace binds funcs to path with TRACE method
func (r *Router) Trace(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodTrace, path, toHandlers(funcs...)...)
}

// Connect binds funcs to path with CONNECT method
func (r *Router) Connect(path string, funcs ...HandlerFunc) {
	r.Bind(http.MethodConnect, path, toHandlers(funcs...)...)
}

func (r *Router) getRoot(method string) *pathpkg.Node {
	root := r.methodToRoot[method]
	if root == nil {
		root = pathpkg.NewEmptyNode()
		r.methodToRoot[method] = root
	}
	return root
}

func (r *Router) createHandlerList(handlers []Handler) *list.List {
	l := list.New()
	for _, h := range r.handlers {
		l.PushBack(h)
	}
	for _, h := range handlers {
		l.PushBack(h)
	}
	return l
}

// Print prints all path trees
func (r *Router) Print() {
	for method, root := range r.methodToRoot {
		nodes := root.ListEndpoints()
		for _, n := range nodes {
			logger.Infof("%-5s %s\t%s", method, n.Path(), handlerListToString(n.Handlers()))
		}
	}
}

func (r *Router) listEndpoints(ctx context.Context, req *Request) Responder {
	l := make(sortableNodeList, 0, 10)
	maxLenOfPath := 0
	nodeToMethod := make(map[*pathpkg.Node]string, 10)
	for method, root := range r.methodToRoot {
		for _, node := range root.ListEndpoints() {
			l = append(l, node)
			nodeToMethod[node] = method
			if n := len(node.Path()); n > maxLenOfPath {
				maxLenOfPath = n
			}
		}
	}
	for _, node := range r.root.ListEndpoints() {
		l = append(l, node)
		nodeToMethod[node] = "*"
		if n := len(node.Path()); n > maxLenOfPath {
			maxLenOfPath = n
		}
	}
	sort.Sort(l)
	b := new(strings.Builder)
	for i, n := range l {
		format := fmt.Sprintf("%%3d. %%6s /%%-%ds %%s\n", maxLenOfPath)
		line := fmt.Sprintf(format, i+1, nodeToMethod[n], n.Path(), handlerListToString(n.Handlers()))
		b.WriteString(line)
	}
	return Text(http.StatusOK, b.String())
}

type sortableNodeList []*pathpkg.Node

func (l sortableNodeList) Len() int {
	return len(l)
}

func (l sortableNodeList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l sortableNodeList) Less(i, j int) bool {
	return strings.Compare(l[i].Path(), l[j].Path()) < 0
}

func equal(v1, v2 interface{}) bool {
	return reflect.TypeOf(v1).Comparable() && reflect.TypeOf(v2).Comparable() && v1 == v2
}
