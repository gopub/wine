package wine

import (
	"context"
	"github.com/gopub/log"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/utils"
)

const defaultMaxRequestMemory = 8 << 20
const defaultRequestTimeout = time.Second * 5
const KeyHTTPResponseWriter = "wine_http_response_writer"
const KeyHTTP2ConnID = "wine_http2_conn_id"

var acceptEncodings = [2]string{"gzip", "defalte"}
var defaultServer *Server
var ShortHandlerNameFlag = true
var Debug = false

// Server implements web server
type Server struct {
	*Router
	*TemplateManager

	Header           http.Header
	MaxRequestMemory int64         //max memory for request, default value is 8M
	RequestTimeout   time.Duration //timeout for each request, default value is 5s
	server           *http.Server

	http2connsToIDs *sync.Map
	idGenerator     types.IDGenerator
}

// NewServer returns a server
func NewServer() *Server {
	s := &Server{
		Router:           NewRouter(),
		TemplateManager:  NewTemplateManager(),
		Header:           make(http.Header),
		MaxRequestMemory: defaultMaxRequestMemory,
		RequestTimeout:   defaultRequestTimeout,
		http2connsToIDs:  &sync.Map{},
	}

	s.Header.Set("Server", "Wine")
	s.AddTemplateFuncMap(template.FuncMap{
		"plus":     plus,
		"minus":    minus,
		"multiple": multiple,
		"divide":   divide,
		"join":     join,
	})

	// 100K ids per second for each shard
	s.idGenerator = types.NewSnakeIDGenerator(0, 10, types.NextMilliseconds, nil, &types.Counter{})
	s.Get("favicon.ico", handleFavIcon)
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

// RunTLS starts server with tls
func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	if s.server != nil {
		log.Panic("Server is running")
	}

	log.Info("Running at", addr, "...")
	s.Router.Print()
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServeTLS(certFile, keyFile)
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
	var http2ConnID int64
	http2Conn := utils.GetHTTP2Conn(rw)
	if http2Conn != nil {
		if idVal, ok := s.http2connsToIDs.Load(http2Conn); ok {
			http2ConnID = idVal.(int64)
		} else {
			http2ConnID = int64(s.idGenerator.NextID())
			s.http2connsToIDs.Store(http2Conn, http2ConnID)
		}
	}

	if !Debug {
		defer func() {
			if e := recover(); e != nil {
				log.Error(e, req)
			}
		}()
	}

	defer func() {
		if http2Conn != nil {
			s.http2connsToIDs.Delete(http2Conn)
		}

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

	if handlers.Empty() {
		handlers = newHandlerList([]Handler{HandlerFunc(handleNotFound)})
	} else {
		handlers.PushBack(HandlerFunc(handleNotImplemented))
	}

	parsedReq := &Request{
		HTTPRequest: req,
		Parameters:  utils.ParseHTTPRequestParameters(req, s.MaxRequestMemory),
	}
	parsedReq.Parameters.AddMapObj(pathParams)

	ctx, cancel := context.WithTimeout(req.Context(), s.RequestTimeout)
	ctx = context.WithValue(ctx, "templates", s.templates)
	defer cancel()

	for k, v := range s.Header {
		rw.Header()[k] = v
	}

	// In case http/2 stream handler needs "responseWriter" to push data to client continuously
	ctx = context.WithValue(ctx, KeyHTTPResponseWriter, rw)
	if http2ConnID > 0 {
		ctx = context.WithValue(ctx, KeyHTTP2ConnID, http2ConnID)
	}

	resp := handlers.Head().Invoke(ctx, parsedReq)
	if resp == nil {
		resp = handleNotImplemented(ctx, parsedReq, nil)
	}
	resp.Respond(ctx, rw)
}

func handleFavIcon(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header()[utils.ContentType] = []string{"image/x-icon"}
		rw.WriteHeader(http.StatusOK)
		rw.Write(_faviconBytes)
	})
}

func handleNotFound(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		log.Warnf("Not found. path=%s", req.HTTPRequest.URL.Path)
		http.Error(rw, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
}

func handleNotImplemented(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		log.Warnf("Not implemented. path=%s", req.HTTPRequest.URL.Path)
		http.Error(rw, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	})
}
