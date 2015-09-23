package wine

import "strings"

type Handler func(*Context) bool

type Routing interface {
	Use(handlers ...Handler)
	Group(relativePath string) Routing
	Bind(method string, path string, handlers ...Handler)
	Match(method string, path string) (handlers []Handler, params map[string]string)
	Get(path string, handlers ...Handler)
	Post(path string, handlers ...Handler)
	Delete(path string, handlers ...Handler)
	Put(path string, handlers ...Handler)
}

type Router struct {
	methodTrees map[string]*node
	basePath    string
	handlers    []Handler
}

func NewRouter() *Router {
	r := &Router{}
	r.methodTrees = make(map[string]*node, 4)
	return r
}

func (this *Router) Group(relativePath string) Routing {
	if len(relativePath) == 0 || relativePath == "/" {
		panic("relative path can not be empty")
	}

	r := &Router{}
	r.methodTrees = this.methodTrees
	r.basePath = cleanPath(this.basePath + "/" + relativePath)
	r.handlers = this.handlers
	return r
}

func (this *Router) Use(handlers ...Handler) {
	this.handlers = append(this.handlers, handlers...)
}

func (this *Router) Match(method string, path string) (handlers []Handler, params map[string]string) {
	n := this.methodTrees[method]
	if n == nil {
		return
	}

	segments := strings.Split(path, "/")
	if len(segments) > 1 && segments[len(segments)-1] == "" {
		segments = segments[0 : len(segments)-1]
	}
	return n.match(segments, path)
}

func (this *Router) Bind(method string, path string, handlers ...Handler) {
	if path == "" {
		panic("path can not be empty")
	}

	if len(method) == 0 {
		panic("method can not be empty")
	}

	if len(handlers) == 0 {
		panic("requre at least one handler")
	}

	n := this.methodTrees[method]
	if n == nil {
		n = &node{}
		n.t = staticNode
		this.methodTrees[method] = n
	}

	path = cleanPath(path)
	if len(this.basePath) == 0 {
		if path == "/" {
			if len(n.handlers) == 0 {
				n.handlers = handlers
			} else {
				panic("path conflicts")
			}
		}
	}

	fullPath := cleanPath(this.basePath + "/" + path)
	hs := append(this.handlers, handlers...)
	segments := strings.Split(fullPath, "/")
	n.addChild(segments[1:], fullPath, hs...)
}

func (this *Router) Get(path string, handlers ...Handler) {
	this.Bind("GET", path, handlers...)
}

func (this *Router) Post(path string, handlers ...Handler) {
	this.Bind("POST", path, handlers...)
}

func (this *Router) Delete(path string, handlers ...Handler) {
	this.Bind("DELETE", path, handlers...)
}

func (this *Router) Put(path string, handlers ...Handler) {
	this.Bind("PUT", path, handlers...)
}
