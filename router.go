package wine

import (
	"net/http"
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

	Match(method, path string) (handlers []Handler, pathParams map[string]string)

	Get(path string, handlers ...HandlerFunc)
	Post(path string, handlers ...HandlerFunc)
	Put(path string, handlers ...HandlerFunc)
	Delete(path string, handlers ...HandlerFunc)
	Any(path string, handlers ...HandlerFunc)
}
