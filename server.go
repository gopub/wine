package wine

import (
	"github.com/justintan/gox/runtime"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Server struct {
	Router
	Header      http.Header
	context     Context
	templates   []*template.Template
	contextPool sync.Pool
}

func NewServer() *Server {
	s := &Server{}
	s.Router = NewDefaultRouter()
	s.Header = make(http.Header)
	s.RegisterContext(&DefaultContext{})
	return s
}

func Default() *Server {
	s := NewServer()
	s.Use(Logger)
	return s
}

func (s *Server) RegisterContext(c Context) {
	if c == nil {
		panic("[WINE] c is nil")
	}
	s.context = c
}

func (s *Server) newContext() interface{} {
	var c Context
	runtime.Renew(&c, s.context)
	return c
}

func (s *Server) Run(addr string) error {
	s.contextPool.New = s.newContext
	log.Println("[WINE] Running at", addr, "...")
	if r, ok := s.Router.(*DefaultRouter); ok {
		r.Print()
	}
	err := http.ListenAndServe(addr, s)
	return err
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("[WINE] ServeHTTP", e, req)
		}
	}()

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}
	path = normalizePath(path)
	handlers, params := s.Match(req.Method, path)
	if len(handlers) == 0 {
		if path == "/favicon.ico/" || path == "/favicon.ico" || path == "favicon.ico" {
			rw.Header()["Content-Type"] = []string{"image/x-icon"}
			rw.WriteHeader(http.StatusOK)
			rw.Write(faviconBytes)
		} else {
			log.Println("[WINE] Not found[", path, "]", req)
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
		return
	}

	c, _ := s.contextPool.Get().(Context)
	c.Rebuild(rw, req, s.templates, handlers)

	c.Params().AddMapObj(params)
	for k, v := range s.Header {
		c.Header()[k] = v
	}
	c.Next()
	if c.Responded() == false {
		c.Status(http.StatusNotFound)
	}
	s.contextPool.Put(c)
}

func (s *Server) AddGlobTemplate(pattern string) {
	tpl := template.Must(template.ParseGlob(pattern))
	s.AddTemplate(tpl)
}

func (s *Server) AddFilesTemplate(files ...string) {
	tpl := template.Must(template.ParseFiles(files...))
	s.AddTemplate(tpl)
}

func (s *Server) AddTemplate(tpl *template.Template) {
	s.templates = append(s.templates, tpl)
}
