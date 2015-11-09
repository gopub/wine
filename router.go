package wine

import (
	"net/http"
	"strings"
)

type Router interface {
	Use(handlers ...Handler) Router
	StaticFile(relativePath, filePath string) Router
	StaticDir(relativePath, filePath string) Router
	StaticFS(relativePath string, fs http.FileSystem) Router
	Bind(method, path string, handlers ...Handler) Router
	Match(method, path string) (handlers []Handler, params map[string]string)
	GET(path string, handlers ...Handler) Router
	POST(path string, handlers ...Handler) Router
	DELETE(path string, handlers ...Handler) Router
	PUT(path string, handlers ...Handler) Router
	HEAD(path string, handlers ...Handler) Router
	PATCH(path string, handlers ...Handler) Router
	OPTIONS(path string, handlers ...Handler) Router
	CONNECT(path string, handlers ...Handler) Router
	TRACE(path string, handlers ...Handler) Router
	GP(path string, handlers ...Handler) Router
	ANY(path string, handlers ...Handler) Router
	Print()
}

type GroupRouter interface {
	Router
	Group(relativePath string) GroupRouter
}

type DefaultRouter struct {
	methodTrees map[string]*node
	basePath    string
	handlers    []Handler
}

func NewDefaultRouter() *DefaultRouter {
	r := &DefaultRouter{}
	r.methodTrees = make(map[string]*node, 4)
	return r
}

func (this *DefaultRouter) Group(relativePath string) GroupRouter {
	if len(relativePath) == 0 {
		panic("relative path can not be empty")
	}

	if relativePath == "/" {
		panic("unnecessary to create group \"/\"")
	}

	r := &DefaultRouter{}
	r.methodTrees = this.methodTrees
	r.basePath = cleanPath(this.basePath + "/" + relativePath)
	r.handlers = make([]Handler, len(this.handlers))
	copy(r.handlers, this.handlers)
	return r
}

func (this *DefaultRouter) Use(handlers ...Handler) Router {
	this.handlers = append(this.handlers, handlers...)
	return this
}

func (this *DefaultRouter) Match(method string, path string) (handlers []Handler, params map[string]string) {
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

func (this *DefaultRouter) Bind(method string, path string, handlers ...Handler) Router {
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
				return this
			} else {
				panic("path conflicts")
			}
		}
	}

	fullPath := cleanPath(this.basePath + "/" + path)
	hs := append(this.handlers, handlers...)
	segments := strings.Split(fullPath, "/")
	n.addChild(segments[1:], fullPath, hs...)
	return this
}

func (this *DefaultRouter) StaticFile(relativePath, filePath string) Router {
	this.GET(relativePath, func(c Context) {
		if c.Written() {
			panic("already written")
		}
		c.SendFile(filePath)
		c.MarkWritten()
	})
	return this
}

func (this *DefaultRouter) StaticDir(relativePath, dirPath string) Router {
	this.StaticFS(relativePath, http.Dir(dirPath))
	return this
}

func (this *DefaultRouter) StaticFS(relativePath string, fs http.FileSystem) Router {
	prefix := cleanPath(this.basePath + "/" + relativePath)
	i := strings.Index(prefix, "*")
	if i > 0 {
		prefix = prefix[:i]
	} else {
		relativePath = cleanPath(relativePath + "/*")
	}

	if prefix[len(prefix)-1] != '/' {
		prefix += "/"
	}

	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	this.GET(relativePath, func(c Context) {
		if c.Written() {
			panic("already written")
		}

		fileServer.ServeHTTP(c.ResponseWriter(), c.HttpRequest())
		c.MarkWritten()
	})
	return this
}

func (this *DefaultRouter) GET(path string, handlers ...Handler) Router {
	this.Bind("GET", path, handlers...)
	return this
}

func (this *DefaultRouter) POST(path string, handlers ...Handler) Router {
	this.Bind("POST", path, handlers...)
	return this
}

func (this *DefaultRouter) DELETE(path string, handlers ...Handler) Router {
	this.Bind("DELETE", path, handlers...)
	return this
}

func (this *DefaultRouter) PUT(path string, handlers ...Handler) Router {
	this.Bind("PUT", path, handlers...)
	return this
}

func (this *DefaultRouter) HEAD(path string, handlers ...Handler) Router {
	this.Bind("HEAD", path, handlers...)
	return this
}

func (this *DefaultRouter) PATCH(path string, handlers ...Handler) Router {
	this.Bind("PATCH", path, handlers...)
	return this
}

func (this *DefaultRouter) OPTIONS(path string, handlers ...Handler) Router {
	this.Bind("OPTIONS", path, handlers...)
	return this
}

func (this *DefaultRouter) CONNECT(path string, handlers ...Handler) Router {
	this.Bind("CONNECT", path, handlers...)
	return this
}

func (this *DefaultRouter) TRACE(path string, handlers ...Handler) Router {
	this.Bind("TRACE", path, handlers...)
	return this
}

func (this *DefaultRouter) GP(path string, handlers ...Handler) Router {
	this.GET(path, handlers...)
	this.POST(path, handlers...)
	return this
}

func (this *DefaultRouter) ANY(path string, handlers ...Handler) Router {
	this.GET(path, handlers...)
	this.POST(path, handlers...)
	this.DELETE(path, handlers...)
	this.PUT(path, handlers...)
	this.HEAD(path, handlers...)
	this.OPTIONS(path, handlers...)
	this.PATCH(path, handlers...)
	this.CONNECT(path, handlers...)
	this.TRACE(path, handlers...)
	return this
}

func (this *DefaultRouter) Print() {
	for m, n := range this.methodTrees {
		n.Print(m, "/")
	}
}
