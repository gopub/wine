package wine

import (
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type Routing interface {
	Use(handlers ...Handler) Routing
	Group(relativePath string) Routing
	StaticFile(relativePath, filePath string) Routing
	StaticDir(relativePath, filePath string) Routing
	StaticFS(relativePath string, fs http.FileSystem) Routing
	Bind(method, path string, handlers ...Handler) Routing
	Match(method, path string) (handlers []Handler, params map[string]string)
	GET(path string, handlers ...Handler) Routing
	POST(path string, handlers ...Handler) Routing
	DELETE(path string, handlers ...Handler) Routing
	PUT(path string, handlers ...Handler) Routing
	HEAD(path string, handlers ...Handler) Routing
	PATCH(path string, handlers ...Handler) Routing
	OPTIONS(path string, handlers ...Handler) Routing
	CONNECT(path string, handlers ...Handler) Routing
	TRACE(path string, handlers ...Handler) Routing
	GP(path string, handlers ...Handler) Routing
	ANY(path string, handlers ...Handler) Routing
	Print()
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
	if len(relativePath) == 0 {
		panic("relative path can not be empty")
	}

	if relativePath == "/" {
		panic("unnecessary to create group \"/\"")
	}

	r := &Router{}
	r.methodTrees = this.methodTrees
	r.basePath = cleanPath(this.basePath + "/" + relativePath)
	r.handlers = make([]Handler, len(this.handlers))
	copy(r.handlers, this.handlers)
	return r
}

func (this *Router) Use(handlers ...Handler) Routing {
	this.handlers = append(this.handlers, handlers...)
	return this
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

func (this *Router) Bind(method string, path string, handlers ...Handler) Routing {
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

func (this *Router) StaticFile(relativePath, filePath string) Routing {
	this.GET(relativePath, func(c Context) {
		if c.Written() {
			panic("already written")
		}
		c.SendFile(filePath)
		c.MarkWritten()
	})
	return this
}

func (this *Router) StaticDir(relativePath, dirPath string) Routing {
	this.StaticFS(relativePath, http.Dir(dirPath))
	return this
}

func (this *Router) StaticFS(relativePath string, fs http.FileSystem) Routing {
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

		fileServer.ServeHTTP(c.ResponseWriter(), c.Request())
		c.MarkWritten()
	})
	return this
}

func (this *Router) GET(path string, handlers ...Handler) Routing {
	this.Bind("GET", path, handlers...)
	return this
}

func (this *Router) POST(path string, handlers ...Handler) Routing {
	this.Bind("POST", path, handlers...)
	return this
}

func (this *Router) DELETE(path string, handlers ...Handler) Routing {
	this.Bind("DELETE", path, handlers...)
	return this
}

func (this *Router) PUT(path string, handlers ...Handler) Routing {
	this.Bind("PUT", path, handlers...)
	return this
}

func (this *Router) HEAD(path string, handlers ...Handler) Routing {
	this.Bind("HEAD", path, handlers...)
	return this
}

func (this *Router) PATCH(path string, handlers ...Handler) Routing {
	this.Bind("PATCH", path, handlers...)
	return this
}

func (this *Router) OPTIONS(path string, handlers ...Handler) Routing {
	this.Bind("OPTIONS", path, handlers...)
	return this
}

func (this *Router) CONNECT(path string, handlers ...Handler) Routing {
	this.Bind("CONNECT", path, handlers...)
	return this
}

func (this *Router) TRACE(path string, handlers ...Handler) Routing {
	this.Bind("TRACE", path, handlers...)
	return this
}

func (this *Router) GP(path string, handlers ...Handler) Routing {
	this.GET(path, handlers...)
	this.POST(path, handlers...)
	return this
}

func (this *Router) ANY(path string, handlers ...Handler) Routing {
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

func (this *Router) Print() {
	for m, n := range this.methodTrees {
		n.Print(m, "/")
	}
}
