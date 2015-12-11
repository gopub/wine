package wine

import (
	"github.com/justintan/gox"
	"html/template"
	"net/http"
	"strings"
)

type Server struct {
	Router
	Header         http.Header
	ContextCreator NewContextFunc
	templates      []*template.Template
}

func NewServer() *Server {
	s := &Server{}
	s.Router = NewDefaultRouter()
	s.Header = make(http.Header)
	s.ContextCreator = NewDefaultContext
	return s
}

func Default() *Server {
	s := &Server{}
	s.Router = NewDefaultRouter()
	s.Header = make(http.Header)
	s.ContextCreator = NewDefaultContext
	s.Use(Logger)
	return s
}

func (s *Server) Run(addr string) error {
	gox.LInfo("Running server at", addr, "...")
	if r, ok := s.Router.(*DefaultRouter); ok {
		r.Print()
	}
	err := http.ListenAndServe(addr, s)
	return err
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			gox.LError("ServeHTTP", e, req)
		}
	}()

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}
	path = cleanPath(path)
	handlers, params := s.Match(req.Method, path)
	if len(handlers) == 0 {
		gox.LError("Not found[", path, "]", req)
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	context := s.ContextCreator(rw, req, s.templates, handlers)
	context.Params().AddMapObj(params)
	for k, v := range s.Header {
		context.Header()[k] = v
	}
	context.Next()
	if context.Responded() == false {
		context.Status(http.StatusNotFound)
	}
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
