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
	Put(path string, handlers ...HandlerFunc)
	Delete(path string, handlers ...HandlerFunc)
	Any(path string, handlers ...HandlerFunc)
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

func (dr *DefaultRouter) Group(relativePath string) Router {
	if len(relativePath) == 0 {
		panic("relative path can not be empty")
	}

	if relativePath == "/" {
		panic("unnecessary to create group \"/\"")
	}

	r := &DefaultRouter{}
	r.methodTrees = dr.methodTrees
	r.basePath = cleanPath(dr.basePath + "/" + relativePath)
	r.handlers = make([]Handler, len(dr.handlers))
	copy(r.handlers, dr.handlers)
	return r
}

func (dr *DefaultRouter) UseHandlers(handlers ...Handler) Router {
	if len(handlers) == 0 {
		return dr
	}
	dr.handlers = append(dr.handlers, handlers...)
	return dr
}

func (dr *DefaultRouter) Use(funcList ...HandlerFunc) Router {
	if len(funcList) == 0 {
		return dr
	}
	return dr.UseHandlers(convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Match(method string, path string) (handlers []Handler, params map[string]string) {
	n := dr.methodTrees[method]
	if n == nil {
		return
	}

	segments := strings.Split(path, "/")
	if len(segments) > 1 && segments[len(segments)-1] == "" {
		segments = segments[0 : len(segments)-1]
	}
	return n.match(segments, path)
}

func (dr *DefaultRouter) Bind(method string, path string, handlers ...Handler) {
	if path == "" {
		panic("path can not be empty")
	}

	if len(method) == 0 {
		panic("method can not be empty")
	}

	if len(handlers) == 0 {
		panic("requre at least one handler")
	}

	n := dr.methodTrees[method]
	if n == nil {
		n = &node{}
		n.t = staticNode
		dr.methodTrees[method] = n
	}

	path = cleanPath(path)
	if len(dr.basePath) == 0 {
		if path == "/" {
			if len(n.handlers) == 0 {
				n.handlers = handlers
				return
			} else {
				panic("path conflicts")
			}
		}
	}

	fullPath := cleanPath(dr.basePath + "/" + path)
	hs := append(dr.handlers, handlers...)
	segments := strings.Split(fullPath, "/")
	n.addChild(segments[1:], fullPath, hs...)
	return
}

func (dr *DefaultRouter) StaticFile(relativePath, filePath string) {
	dr.Get(relativePath, func(c Context) {
		c.File(filePath)
	})
	return
}

func (dr *DefaultRouter) StaticDir(relativePath, dirPath string) {
	dr.StaticFS(relativePath, http.Dir(dirPath))
	return
}

func (dr *DefaultRouter) StaticFS(relativePath string, fs http.FileSystem) {
	prefix := cleanPath(dr.basePath + "/" + relativePath)
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
	dr.Get(relativePath, func(c Context) {
		c.ServeHTTP(fileServer)
	})
	return
}

func (dr *DefaultRouter) HandleGet(path string, handlers ...Handler) {
	dr.Bind("GET", path, handlers...)
	return
}

func (dr *DefaultRouter) HandlePost(path string, handlers ...Handler) {
	dr.Bind("POST", path, handlers...)
	return
}

func (dr *DefaultRouter) HandlePut(path string, handlers ...Handler) {
	dr.Bind("PUT", path, handlers...)
	return
}

func (dr *DefaultRouter) HandleDelete(path string, handlers ...Handler) {
	dr.Bind("DELETE", path, handlers...)
	return
}

func (dr *DefaultRouter) HandleAny(path string, handlers ...Handler) {
	dr.HandleGet(path, handlers...)
	dr.HandlePost(path, handlers...)
	dr.HandlePut(path, handlers...)
	return
}

func (dr *DefaultRouter) Get(path string, funcList ...HandlerFunc) {
	dr.HandleGet(path, convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Post(path string, funcList ...HandlerFunc) {
	dr.HandlePost(path, convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Put(path string, funcList ...HandlerFunc) {
	dr.HandlePut(path, convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Delete(path string, funcList ...HandlerFunc) {
	dr.HandleDelete(path, convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Any(path string, funcList ...HandlerFunc) {
	dr.HandleAny(path, convertToHandlers(funcList...)...)
}

func (dr *DefaultRouter) Print() {
	for m, n := range dr.methodTrees {
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
