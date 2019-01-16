package wine

import (
	"context"
	"github.com/gopub/log"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gopub/utils"
)

const defaultMaxRequestMemory = 1 << 20
const defaultRequestTimeout = time.Second * 5
const keyHTTPResponseWriter = "wine_http_response_writer"
const keyHTTP2ConnID = "wine_http2_conn_id"
const keyTemplates = "wine_templates"

var acceptEncodings = [2]string{"gzip", "defalte"}
var ShortHandlerNameFlag = true

// Server implements web server
type Server struct {
	*Router
	*TemplateManager

	Header           http.Header
	MaxRequestMemory int64         //max memory for request, default value is 8M
	RequestTimeout   time.Duration //timeout for each request, default value is 5s
	server           *http.Server

	h2connToIDMu sync.RWMutex
	h2connToID   map[interface{}]string

	RequestParser RequestParser
}

// NewServer returns a server
func NewServer(config *Config) *Server {
	s := &Server{
		Router:           NewRouter(),
		TemplateManager:  NewTemplateManager(),
		Header:           make(http.Header),
		MaxRequestMemory: defaultMaxRequestMemory,
		RequestTimeout:   defaultRequestTimeout,
		h2connToID:       make(map[interface{}]string, 1024),
		RequestParser:    NewDefaultRequestParser(),
	}

	s.Header.Set("Server", "Wine")
	s.AddTemplateFuncMap(template.FuncMap{
		"plus":     plus,
		"minus":    minus,
		"multiple": multiple,
		"divide":   divide,
		"join":     join,
	})

	if config != nil {
		s.handlers = config.Handlers
	}

	s.Get("favicon.ico", handleFavIcon)
	return s
}

// Run starts server
func (s *Server) Run(addr string) error {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	logger.Info("Running at", addr, "...")
	s.Router.Print()
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	if err != nil {
		logger.Error(err)
	}
	return err
}

// RunTLS starts server with tls
func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	logger.Info("Running at", addr, "...")
	s.Router.Print()
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		logger.Error(err)
	}
	return err
}

// Shutdown stops server
func (s *Server) Shutdown() {
	s.server.Shutdown(context.Background())
	logger.Info("Shutdown")
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	startAt := time.Now()
	h2conn, h2connID := s.detectHTTP2(rw)
	if logger.Level() <= log.DebugLevel {
		defer func() {
			if e := recover(); e != nil {
				logger.Error(e, req)
			}
		}()
	}

	defer func() {
		if h2conn != nil {
			s.h2connToIDMu.Lock()
			delete(s.h2connToID, h2conn)
			s.h2connToIDMu.Unlock()
		}

		var statGetter statusGetter
		if cw, ok := rw.(*compressedResponseWriter); ok {
			cw.Close()
			statGetter = cw.ResponseWriter.(statusGetter)
		}

		if statGetter == nil {
			statGetter = rw.(statusGetter)
		}

		if statGetter.Status() >= 400 {
			logger.Errorf("%s %s %s %d %v : request=%v",
				req.RemoteAddr,
				req.Method,
				req.RequestURI,
				statGetter.Status(),
				time.Since(startAt), req)
		} else {
			logger.Infof("%s %s %s %d %v",
				req.RemoteAddr,
				req.Method,
				req.RequestURI,
				statGetter.Status(),
				time.Since(startAt))
		}
	}()

	// Add compression to responseWriter
	rw = &responseWriterWrapper{ResponseWriter: rw}
	rw = wrapperCompressedWriter(rw, req)

	path := getRequestPath(req)
	method := strings.ToUpper(req.Method)
	handlers, pathParams := s.Match(method, path)

	if handlers.Empty() {
		handlers = newHandlerList([]Handler{HandlerFunc(handleNotFound)})
	} else {
		handlers.PushBack(HandlerFunc(handleNotImplemented))
	}

	params, err := s.RequestParser.ParseHTTPRequest(req, s.MaxRequestMemory)

	if err != nil {
		return
	}
	params.AddMapObj(pathParams)

	parsedReq := &Request{
		HTTPRequest: req,
		Parameters:  params,
	}

	ctx, cancel := context.WithTimeout(req.Context(), s.RequestTimeout)
	defer cancel()

	for k, v := range s.Header {
		rw.Header()[k] = v
	}

	ctx = context.WithValue(ctx, keyTemplates, s.templates)
	// In case http/2 stream handler needs "responseWriter" to push data to client continuously
	ctx = context.WithValue(ctx, keyHTTPResponseWriter, rw)
	if len(h2connID) > 0 {
		ctx = context.WithValue(ctx, keyHTTP2ConnID, h2connID)
	}

	resp := handlers.Head().Invoke(ctx, parsedReq)

	if resp == nil {
		resp = handleNotImplemented(ctx, parsedReq, nil)
	}
	resp.Respond(ctx, rw)
}

func (s *Server) detectHTTP2(rw http.ResponseWriter) (conn interface{}, connID string) {
	h2conn := GetHTTP2Conn(rw)
	if h2conn == nil {
		return nil, ""
	}

	// Hope RLock is faster than Lock?
	s.h2connToIDMu.RLock()
	h2connID, ok := s.h2connToID[h2conn]
	s.h2connToIDMu.RUnlock()
	if !ok {
		s.h2connToIDMu.Lock()
		h2connID, ok = s.h2connToID[h2conn]
		if !ok {
			h2connID = utils.UniqueID()
			s.h2connToID[h2conn] = h2connID
			logger.Debug("New http/2 conn:", h2connID, h2conn)
		}
		s.h2connToIDMu.Unlock()
	}

	return h2conn, h2connID
}

func wrapperCompressedWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	// Add compression to responseWriter
	ae := req.Header.Get("Accept-Encoding")
	for _, enc := range acceptEncodings {
		if strings.Contains(ae, enc) {
			rw.Header().Set("Content-Encoding", enc)
			if cw, err := newCompressedResponseWriter(rw, enc); err == nil {
				return cw
			}
		}
	}

	return rw
}

func getRequestPath(req *http.Request) string {
	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}

	return normalizePath(path)
}

func handleFavIcon(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header()[ContentType] = []string{"image/x-icon"}
		rw.WriteHeader(http.StatusOK)
		rw.Write(_faviconBytes)
	})
}

func handleNotFound(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		log.Warnf("%s %s %d %s", req.HTTPRequest.Method, req.HTTPRequest.URL.Path,
			http.StatusNotFound, http.StatusText(http.StatusNotFound))
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

func handleNotImplemented(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		log.Warnf("%s %s %d %s", req.HTTPRequest.Method, req.HTTPRequest.URL.Path,
			http.StatusNotFound, http.StatusText(http.StatusNotFound))
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	})
}

func GetHTTP2ConnID(ctx context.Context) string {
	id, _ := ctx.Value(keyHTTP2ConnID).(string)
	return id
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(keyHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func GetTemplates(ctx context.Context) []*template.Template {
	v, _ := ctx.Value(keyTemplates).([]*template.Template)
	return v
}
