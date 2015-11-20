package wine

import (
	"net/http"
	"strings"
)

type Router interface {
	Group(relativePath string) Router
	UseHandlers(handlers ...Handler) Router
	Use(funcList ...HandlerFunc) Router
	StaticFile(relativePath, filePath string)
	StaticDir(relativePath, filePath string)
	StaticFS(relativePath string, fs http.FileSystem)
	Bind(method, path string, handlers ...Handler)
	Match(method, path string) (handlers []Handler, params map[string]string)

	HandleGet(path string, handlers ...Handler)
	HandlePost(path string, handlers ...Handler)
	HandleDelete(path string, handlers ...Handler)
	HandlePut(path string, handlers ...Handler)
	HandleAny(path string, handlers ...Handler)

	Get(path string, handlers ...HandlerFunc)
	Post(path string, handlers ...HandlerFunc)
	Delete(path string, handlers ...HandlerFunc)
	Put(path string, handlers ...HandlerFunc)
	Any(path string, handlers ...HandlerFunc)

	Print()
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

func (this *DefaultRouter) Group(relativePath string) Router {
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

func (this *DefaultRouter) UseHandlers(handlers ...Handler) Router {
	if len(handlers) == 0 {
		return this
	}
	this.handlers = append(this.handlers, handlers...)
	return this
}

func (this *DefaultRouter) Use(funcList ...HandlerFunc) Router {
	if len(funcList) == 0 {
		return this
	}
	return this.UseHandlers(convertToHandlers(funcList...)...)
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

func (this *DefaultRouter) Bind(method string, path string, handlers ...Handler) {
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
				return
			} else {
				panic("path conflicts")
			}
		}
	}

	fullPath := cleanPath(this.basePath + "/" + path)
	hs := append(this.handlers, handlers...)
	segments := strings.Split(fullPath, "/")
	n.addChild(segments[1:], fullPath, hs...)
	return
}

func (this *DefaultRouter) StaticFile(relativePath, filePath string) {
	this.Get(relativePath, func(c Context) {
		c.File(filePath)
	})
	return
}

func (this *DefaultRouter) StaticDir(relativePath, dirPath string) {
	this.StaticFS(relativePath, http.Dir(dirPath))
	return
}

func (this *DefaultRouter) StaticFS(relativePath string, fs http.FileSystem) {
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
	this.Get(relativePath, func(c Context) {
		c.ServeHTTP(fileServer)
	})
	return
}

func (this *DefaultRouter) HandleGet(path string, handlers ...Handler) {
	this.Bind("GET", path, handlers...)
	return
}

func (this *DefaultRouter) HandlePost(path string, handlers ...Handler) {
	this.Bind("POST", path, handlers...)
	return
}

func (this *DefaultRouter) HandleDelete(path string, handlers ...Handler) {
	this.Bind("DELETE", path, handlers...)
	return
}

func (this *DefaultRouter) HandlePut(path string, handlers ...Handler) {
	this.Bind("PUT", path, handlers...)
	return
}

func (this *DefaultRouter) HandleAny(path string, handlers ...Handler) {
	this.HandleGet(path, handlers...)
	this.HandlePost(path, handlers...)
	this.HandleDelete(path, handlers...)
	this.HandlePut(path, handlers...)
	return
}

func (this *DefaultRouter) Get(path string, funcList ...HandlerFunc) {
	this.HandleGet(path, convertToHandlers(funcList...)...)
}

func (this *DefaultRouter) Post(path string, funcList ...HandlerFunc) {
	this.HandlePost(path, convertToHandlers(funcList...)...)
}

func (this *DefaultRouter) Delete(path string, funcList ...HandlerFunc) {
	this.HandleDelete(path, convertToHandlers(funcList...)...)
}

func (this *DefaultRouter) Put(path string, funcList ...HandlerFunc) {
	this.HandlePut(path, convertToHandlers(funcList...)...)
}

func (this *DefaultRouter) Any(path string, funcList ...HandlerFunc) {
	this.HandleAny(path, convertToHandlers(funcList...)...)
}

func (this *DefaultRouter) Print() {
	for m, n := range this.methodTrees {
		n.Print(m, "/")
	}
}

func convertToHandlers(funcList ...HandlerFunc) []Handler {
	handlers := make([]Handler, len(funcList))
	for i, h := range funcList {
		handlers[i] = h
	}
	return handlers
}
