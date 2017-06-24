package wine

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/natande/gox"
)

const _DefaultMaxRequestMemory = 8 << 20

var _acceptEncodings = [2]string{"gzip", "defalte"}

// Server implements web server
type Server struct {
	Router
	Header           http.Header
	MaxRequestMemory int64 //max memory for request, default value is 8M
	responder        Responder
	templates        []*template.Template
	templateFuncs    template.FuncMap
	contextPool      sync.Pool
	server           *http.Server
}

// NewServer returns a server
func NewServer() *Server {
	s := &Server{}
	s.Router = NewDefaultRouter()
	s.responder = &DefaultResponder{}
	s.MaxRequestMemory = _DefaultMaxRequestMemory
	s.Header = make(http.Header)
	s.Header.Set("Server", "Wine")
	s.AddTemplateFuncs(template.FuncMap{
		"plus":     plus,
		"minus":    minus,
		"multiple": multiple,
		"divide":   divide,
		"join":     join,
	})
	return s
}

// DefaultServer returns a default server with Logger interceptor
func DefaultServer() *Server {
	s := NewServer()
	s.Use(Logger)
	return s
}

func (s *Server) RegisterResponder(r Responder) {
	if r == nil {
		panic("[WINE] r is nil")
	}
	s.responder = r
}

func (s *Server) newContext() interface{} {
	c := &Context{}
	c.keyValues = gox.M{}
	gox.Renew(&c.Responder, s.responder)
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
	if err != nil {
		log.Println("[WINE]", err)
	}
	return err
}

// Shutdown stops server
func (s *Server) Shutdown() {
	s.server.Shutdown(context.Background())
	log.Println("[WINE] Shutdown")
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
	ae := req.Header.Get("Accept-Encoding")
	for _, enc := range _acceptEncodings {
		if strings.Contains(ae, enc) {
			rw.Header().Set("Content-Encoding", enc)
			if cw, err := newCompressedResponseWriter(rw, enc); err == nil {
				rw = cw
			}
			break
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
			rw.Write(_faviconBytes)
		} else {
			log.Println("[WINE] Not found[", path, "]", req)
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
		return
	}

	c := s.makeContext(rw, req, handlers)

	c.Params().AddMapObj(pathParams)
	c.Next()
	if !c.Responded() {
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

func (s *Server) makeContext(rw http.ResponseWriter, req *http.Request, handlers []Handler) *Context {
	c := s.contextPool.Get().(*Context)
	if c.keyValues != nil {
		for k := range c.keyValues {
			delete(c.keyValues, k)
		}
	} else {
		c.keyValues = gox.M{}
	}

	c.Responder.Reset(req, rw, s.templates)
	c.req = req
	c.reqParams = gox.ParseHTTPRequestParameters(req, s.MaxRequestMemory)
	if c.handlers == nil {
		c.handlers = NewHandlerChain(handlers)
	} else {
		c.handlers.handlers = handlers
		c.handlers.index = 0
	}

	// Set global headers
	for k, v := range s.Header {
		c.Responder.Header()[k] = v
	}

	return c
}
