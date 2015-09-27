package wine

import (
	"fmt"
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type server struct {
	Routing
	Header http.Header
}

func Server() *server {
	s := &server{}
	s.Routing = NewRouter()
	s.Header = make(http.Header)
	return s
}

func (this *server) Run(addr string) error {
	gox.LInfo("Running at", addr, "...")
	this.Print()
	err := http.ListenAndServe(addr, this)
	return err
}

func (this *server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			gox.LError("ServeHTTP", e)
		}
	}()

	gox.LCritical(fmt.Sprintf("%s %s %q", req.Method, req.Header.Get(gox.ContentTypeName), req.RequestURI))

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}

	handlers, params := this.Match(req.Method, path)
	if len(handlers) == 0 {
		gox.LError("No service ", path, "[", req.RequestURI, "]")
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := NewContext(rw, req, handlers, params, this.Header)
	context.Next()
	if context.Written() == false {
		context.Status(http.StatusNotFound)
	}
}
