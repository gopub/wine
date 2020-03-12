package wine

import (
	"container/list"
	"context"
	"fmt"
	"github.com/gopub/gox"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gopub/wine/internal/debug"

	pathpkg "github.com/gopub/wine/internal/path"
)

// Router implements routing function
type Router struct {
	anyRoot      *pathpkg.Node // Root for any methods
	methodToRoot map[string]*pathpkg.Node
	basePath     string
	handlers     []Handler
}

// NewRouter new a Router
func NewRouter() *Router {
	r := &Router{
		anyRoot:      pathpkg.NewEmptyNode(),
		methodToRoot: make(map[string]*pathpkg.Node, 4),
	}
	r.bindSysHandlers()
	return r
}

func (r *Router) bindSysHandlers() {
	r.Get(endpointPath, r.listEndpoints)
	r.Get(sysDatePath, handleDate)
	r.Any(echoPath, handleEcho)
	if h, ok := debug.ByteStreamHandler.(Handler); ok {
		r.Bind(http.MethodGet, byteStreamPath, h)
	}
	if h, ok := debug.TextStreamHandler.(Handler); ok {
		r.Bind(http.MethodGet, textStreamPath, h)
	}
	if h, ok := debug.JSONStreamHandler.(Handler); ok {
		r.Bind(http.MethodGet, jsonStreamPath, h)
	}
}

func (r *Router) clone() *Router {
	nr := &Router{
		anyRoot:      r.anyRoot,
		methodToRoot: r.methodToRoot,
		basePath:     r.basePath,
	}
	nr.handlers = make([]Handler, len(r.handlers))
	copy(nr.handlers, r.handlers)
	return nr
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
		found := false
		for _, rh := range nr.handlers {
			if gox.Equal(h, rh) {
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
func (r *Router) Use(funcList ...HandlerFunc) *Router {
	return r.UseHandlers(toHandlers(funcList...)...)
}

// match finds handlers and parses path parameters according to method and path
func (r *Router) match(method string, path string) (*list.List, map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}

	root := r.methodToRoot[method]
	if root == nil {
		root = r.anyRoot
	}

	m, params := root.Match(segments...)
	if m == nil && root != r.anyRoot {
		m, params = r.anyRoot.Match(segments...)
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

// Bind binds method, path with handlers
func (r *Router) Bind(method, path string, handlers ...Handler) {
	if path == "" {
		logger.Panic("Empty path")
	}

	if len(method) == 0 {
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
		if r.anyRoot.IsEndpoint() {
			logger.Panicf("Conflict: ANY, %s", r.basePath)
		} else if root.IsEndpoint() {
			logger.Panicf("Conflict: %s, %s", method, r.basePath)
		} else {
			root.SetHandlers(hl)
		}
	} else {
		nodeList := pathpkg.NewNodeList(path, hl)
		if pair := r.anyRoot.Conflict(nodeList); pair != nil {
			first := pair.First.(*pathpkg.Node).Path()
			second := pair.Second.(*pathpkg.Node).Path()
			logger.Panicf("Conflict: ANY %s, %s %s", first, method, second)
		}
		root.Add(nodeList)
	}
}

// StaticFile binds path to a file
func (r *Router) StaticFile(path, filePath string) {
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responder {
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
	if len(prefix) == 0 {
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
	r.Get(path, func(ctx context.Context, req *Request, next Invoker) Responder {
		return Handle(req.request, fileServer)
	})
}

// Get binds funcList to path with GET method
func (r *Router) Any(path string, funcList ...HandlerFunc) {
	if path == "" {
		logger.Panic("Empty path")
	}
	handlers := toHandlers(funcList...)
	if len(handlers) == 0 {
		logger.Panic("No handlers")
	}

	hl := r.createHandlerList(handlers)
	path = pathpkg.Normalize(r.basePath + "/" + path)
	if path == "" {
		if r.anyRoot.IsEndpoint() {
			logger.Panicf("Conflict: ANY, %s", r.basePath)
		} else {
			r.anyRoot.SetHandlers(hl)
		}
	} else {
		nodeList := pathpkg.NewNodeList(path, hl)
		r.anyRoot.Add(nodeList)
	}
}

// Get binds funcList to path with GET method
func (r *Router) Get(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodGet, path, toHandlers(funcList...)...)
}

// Post binds funcList to path with POST method
func (r *Router) Post(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodPost, path, toHandlers(funcList...)...)
}

// Put binds funcList to path with PUT method
func (r *Router) Put(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodPut, path, toHandlers(funcList...)...)
}

// Patch binds funcList to path with PATCH method
func (r *Router) Patch(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodPatch, path, toHandlers(funcList...)...)
}

// Delete binds funcList to path with DELETE method
func (r *Router) Delete(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodDelete, path, toHandlers(funcList...)...)
}

// Options binds funcList to path with OPTIONS method
func (r *Router) Options(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodOptions, path, toHandlers(funcList...)...)
}

// Head binds funcList to path with HEAD method
func (r *Router) Head(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodHead, path, toHandlers(funcList...)...)
}

// Trace binds funcList to path with TRACE method
func (r *Router) Trace(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodTrace, path, toHandlers(funcList...)...)
}

// Connect binds funcList to path with CONNECT method
func (r *Router) Connect(path string, funcList ...HandlerFunc) {
	r.Bind(http.MethodConnect, path, toHandlers(funcList...)...)
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

func (r *Router) listEndpoints(ctx context.Context, req *Request, next Invoker) Responder {
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
	for _, node := range r.anyRoot.ListEndpoints() {
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
