package wine

import (
	"github.com/justintan/gox"
	"net/http"
	"strings"
)

type server struct {
	Routing
	Header     http.Header
	NewContext NewContextFunc
}

func Server() *server {
	s := &server{}
	s.Routing = NewRouter()
	s.Header = make(http.Header)
	s.NewContext = NewDefaultContext
	return s
}

func Default() *server {
	s := &server{}
	s.Routing = NewRouter()
	s.Header = make(http.Header)
	s.NewContext = NewDefaultContext
	s.Use(Logger)
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

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}
	path = cleanPath(path)
	handlers, params := this.Match(req.Method, path)
	if len(handlers) == 0 {
		gox.LError("Not found", path, "[", req.RequestURI, "]")
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := this.NewContext(rw, req, handlers)
	context.RequestParams().AddMap(params)
	for k, v := range this.Header {
		context.RequestHeader()[k] = v
	}
	context.Next()
	if context.Written() == false {
		context.SendStatus(http.StatusNotFound)
	}
}
