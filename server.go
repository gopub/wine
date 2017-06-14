package wine

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/justintan/gox/runtime"
)

// Server implements web server
type Server struct {
	Router
	Header        http.Header
	context       Context
	templates     []*template.Template
	templateFuncs template.FuncMap
	contextPool   sync.Pool
	server        *http.Server
}

// NewServer returns a server
func NewServer() *Server {
	s := &Server{}
	s.Router = NewDefaultRouter()
	s.Header = make(http.Header)
	s.Header.Set("Server", "Wine")
	s.RegisterContext(&DefaultContext{})
	s.AddTemplateFuncs(template.FuncMap{
		"plus":     plus,
		"minus":    minus,
		"multiple": multiple,
		"divide":   divide,
		"join":     join,
	})
	return s
}

// Default returns a default server with Logger interceptor
func Default() *Server {
	s := NewServer()
	s.Use(Logger)
	return s
}

// RegisterContext registers a Context implementation
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

// Run starts server
func (s *Server) Run(addr string) error {
	if s.server != nil {
		panic("[WINE] Server is running")
	}

	s.contextPool.New = s.newContext
	log.Println("[WINE] Running at", addr, "...")
	if r, ok := s.Router.(*DefaultRouter); ok {
		r.Print()
	}
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	return err
}

// Shutdown stops server
func (s *Server) Shutdown() {
	s.server.Shutdown(context.Background())
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("[WINE] ServeHTTP", e, req)
		} else {
			if cw, ok := rw.(*compressedResponseWriter); ok {
				cw.Close()
			}
		}
	}()

	// Add compression to responseWriter
	acceptEncoding := req.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "gzip") {
		rw.Header().Set("Content-Encoding", "gzip")
		if cw, err := newCompressedResponseWriter(rw, "gzip"); err == nil {
			rw = cw
		}
	} else if strings.Contains(acceptEncoding, "deflate") {
		rw.Header().Set("Content-Encoding", "deflate")
		if cw, err := newCompressedResponseWriter(rw, "deflate"); err == nil {
			rw = cw
		}
	}

	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}
	path = normalizePath(path)
	method := strings.ToUpper(req.Method)
	handlers, pathParams := s.Match(method, path)
	if len(handlers) == 0 {
		if path == "favicon.ico" {
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

	c.Params().AddMapObj(pathParams)
	// Set global headers
	for k, v := range s.Header {
		c.Header()[k] = v
	}

	c.Next()

	if c.Responded() == false {
		c.Status(http.StatusNotFound)
	}

	s.contextPool.Put(c)
}

// AddGlobTemplate adds a template by parsing template files with pattern
func (s *Server) AddGlobTemplate(pattern string) {
	tmpl := template.Must(template.ParseGlob(pattern))
	s.AddTemplate(tmpl)
}

// AddFilesTemplate adds a template by parsing template files
func (s *Server) AddFilesTemplate(files ...string) {
	tmpl := template.Must(template.ParseFiles(files...))
	s.AddTemplate(tmpl)
}

// AddTextTemplate adds a template by parsing texts
func (s *Server) AddTextTemplate(name string, texts ...string) {
	tmpl := template.New(name)
	for _, txt := range texts {
		tmpl = template.Must(tmpl.Parse(txt))
	}
	s.AddTemplate(tmpl)
}

// AddTemplate adds a template
func (s *Server) AddTemplate(tmpl *template.Template) {
	if s.templateFuncs != nil {
		tmpl.Funcs(s.templateFuncs)
	}
	s.templates = append(s.templates, tmpl)
}

// AddTemplateFuncs adds template functions
func (s *Server) AddTemplateFuncs(funcs template.FuncMap) {
	if funcs == nil {
		panic("funcs is nil")
	}

	if s.templateFuncs == nil {
		s.templateFuncs = funcs
	} else {
		for name, f := range funcs {
			s.templateFuncs[name] = f
		}
	}

	for _, tmpl := range s.templates {
		tmpl.Funcs(funcs)
	}
}
