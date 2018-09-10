package wine

import (
	"context"
	"github.com/gopub/log"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/utils"
)

const defaultMaxRequestMemory = 8 << 20
const defaultRequestTimeout = time.Second * 5

var acceptEncodings = [2]string{"gzip", "defalte"}
var defaultServer *Server

// Server implements web server
type Server struct {
	*Router
	Header           http.Header
	MaxRequestMemory int64         //max memory for request, default value is 8M
	RequestTimeout   time.Duration //timeout for each request, default value is 5s
	templates        []*template.Template
	templateFuncs    template.FuncMap
	server           *http.Server
}

// NewServer returns a server
func NewServer() *Server {
	s := &Server{}
	s.Router = NewRouter()
	s.MaxRequestMemory = defaultMaxRequestMemory
	s.RequestTimeout = defaultRequestTimeout
	s.Header = make(http.Header)
	s.Header.Set("Server", "Wine")
	s.AddTemplateFuncMap(template.FuncMap{
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
	if defaultServer == nil {
		defaultServer = NewServer()
		defaultServer.Use(Logger)
	}
	return defaultServer
}

// Run starts server
func (s *Server) Run(addr string) error {
	if s.server != nil {
		log.Panic("Server is running")
	}

	log.Info("Running at", addr, "...")
	s.Router.Print()
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	if err != nil {
		log.Error(err)
	}
	return err
}

// Shutdown stops server
func (s *Server) Shutdown() {
	s.server.Shutdown(context.Background())
	log.Info("Shutdown")
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			log.Error(e, req)
		}
	}()

	defer func() {
		if cw, ok := rw.(*compressedResponseWriter); ok {
			cw.Close()
		}
	}()

	// Add compression to responseWriter
	ae := req.Header.Get("Accept-Encoding")
	for _, enc := range acceptEncodings {
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
			log.Warnf("Not found. path=%s, request=%v", path, req)
			http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
		return
	}

	request, responder := s.makePair(rw, req, handlers)
	ctx, cancel := context.WithTimeout(req.Context(), s.RequestTimeout)
	defer cancel()
	request.Parameters().AddMapObj(pathParams)
	if !responder.Next(ctx, request, responder) {
		responder.Status(http.StatusNotFound)
	}
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
func (s *Server) AddTemplateFuncMap(funcMap template.FuncMap) {
	if funcMap == nil {
		log.Panic("funcMap is nil")
	}

	if s.templateFuncs == nil {
		s.templateFuncs = funcMap
	} else {
		for name, f := range funcMap {
			s.templateFuncs[name] = f
		}
	}

	for _, tmpl := range s.templates {
		tmpl.Funcs(funcMap)
	}
}

func (s *Server) makePair(rw http.ResponseWriter, req *http.Request, handlers []Handler) (Request, Responder) {
	request := &requestImpl{
		req:       req,
		reqParams: utils.ParseHTTPRequestParameters(req, s.MaxRequestMemory),
		keyValues: types.M{},
	}

	responder := &responderImpl{
		handlers:  newHandlerChain(handlers),
		req:       req,
		writer:    rw,
		templates: s.templates,
	}

	// Set global headers
	for k, v := range s.Header {
		responder.Header()[k] = v
	}

	return request, responder
}
