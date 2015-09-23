package wine

import (
	"fmt"
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type server struct {
	Routing
}

func Server() *server {
	s := &server{}
	s.Routing = NewRouter()
	return s
}

func (this *server) Run(addr string) error {
	err := http.ListenAndServe(addr, this)
	return err
}

func (this *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			gox.Log().Error("ServeHTTP", e)
		}
		gox.Log().Critical(fmt.Sprintf("Handled Request %q", req.RequestURI))
	}()

	gox.Log().Critical(fmt.Sprintf("%s %s %q", req.Method, req.Header.Get(gox.ContentTypeName), req.RequestURI))

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}

	handlers, params := this.Match(req.Method, path)
	if len(handlers) == 0 {
		gox.Log().Error("No service ", path, "[", req.RequestURI, "]")
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := NewContext(rw, req)
	context.Params.AddMap(params)
	for _, h := range handlers {
		if h(context) == false {
			break
		}
	}

	if context.Written() == false {
		context.Status(http.StatusNotFound)
	}
}
