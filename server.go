package wine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gopub/gox"
	"github.com/gopub/gox/env"
	"github.com/gopub/log"
	"github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/template"
)

var acceptEncodings = []string{"gzip", "deflate"}

// ShortHandlerNameFlag means using short name format
var ShortHandlerNameFlag = true

const (
	sysDatePath    = "_sys/date"
	endpointPath   = "_debug/endpoints"
	echoPath       = "_debug/echo"
	byteStreamPath = "_debug/bytestream"
	textStreamPath = "_debug/textstream"
	jsonStreamPath = "_debug/jsonstream"
	faviconPath    = "favicon.ico"
)

var reservedPaths = map[string]bool{
	sysDatePath:    true,
	endpointPath:   true,
	faviconPath:    true,
	echoPath:       true,
	byteStreamPath: true,
	jsonStreamPath: true,
	textStreamPath: true,
}

const (
	defaultSessionTTL = 30 * time.Minute
	minSessionTTL     = 5 * time.Minute
)

// Server implements web server
type Server struct {
	*Router
	*templateManager
	server      *http.Server
	sessionTTL  time.Duration
	sessionName string

	maxRequestMemory   gox.ByteUnit
	Header             http.Header
	Timeout            time.Duration //Timeout for each request, default value is 20s
	BeginHandler       Handler
	CompressionEnabled bool
	Recovery           bool

	invokers struct {
		favicon  *invokerList
		notfound *invokerList
		options  *invokerList
	}

	logger *log.Logger
}

// NewServer returns a server
func NewServer() *Server {
	logger := log.GetLogger("Wine")
	logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)
	header := make(http.Header, 1)
	header.Set("Server", "Wine")

	s := &Server{
		sessionName:        env.String("wine.session.name", "wsessionid"),
		sessionTTL:         env.Duration("wine.session.ttl", defaultSessionTTL),
		maxRequestMemory:   gox.ByteUnit(env.SizeInBytes("wine.max_memory", int(8*gox.MB))),
		Router:             NewRouter(),
		templateManager:    newTemplateManager(),
		Header:             header,
		Timeout:            env.Duration("wine.timeout", 10*time.Second),
		logger:             logger,
		CompressionEnabled: env.Bool("wine.compression", true),
		Recovery:           env.Bool("wine.recovery", true),
	}
	if s.sessionTTL < minSessionTTL {
		s.sessionTTL = minSessionTTL
	}
	s.invokers.favicon = newInvokerList(toHandlerList(HandlerFunc(handleFavIcon)))
	s.invokers.notfound = newInvokerList(toHandlerList(HandlerFunc(handleNotFound)))
	s.invokers.options = newInvokerList(toHandlerList(HandlerFunc(s.handleOptions)))
	s.AddTemplateFuncMap(template.FuncMap)
	return s
}

// Run starts server
func (s *Server) Run(addr string) {
	if s.server != nil {
		logger.Fatalf("Server is running")
	}

	logger.Info("Running at", addr, "...")
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Server closed")
		} else {
			logger.Fatalf("ListenAndServe: %v", err)
		}
	}
}

// RunTLS starts server with tls
func (s *Server) RunTLS(addr, certFile, keyFile string) {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	logger.Infof("Running at %s ...", addr)
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Server closed")
		} else {
			logger.Fatalf("ListenAndServe: %v", err)
		}
	}
}

// Shutdown stops server
func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if s.Recovery {
		defer func() {
			if e := recover(); e != nil {
				logger.Errorf("%v: %+v\n", req, e)
				logger.Errorf("\n%s\n", string(debug.Stack()))
			}
		}()
	}
	rw = s.wrapResponseWriter(rw, req)
	defer s.closeWriter(rw)
	defer s.logHTTP(rw, req, time.Now())

	sid := s.initSession(rw, req)
	ctx, cancel := s.setupContext(req.Context(), rw, sid)
	defer cancel()

	parsedReq, err := parseRequest(req, s.maxRequestMemory)
	if err != nil {
		logger.Errorf("Parse request: %v", err)
		resp := Text(http.StatusBadRequest, fmt.Sprintf("Parse request: %v", err))
		resp.Respond(ctx, rw)
		return
	}
	parsedReq.params[s.sessionName] = sid
	s.serve(ctx, parsedReq, rw)
}

func (s *Server) serve(ctx context.Context, req *Request, rw http.ResponseWriter) {
	path := req.NormalizedPath()
	method := strings.ToUpper(req.Request().Method)
	var invokers *invokerList
	handlers, params := s.match(method, path)
	req.params.AddMapObj(params)
	if handlers != nil && handlers.Len() > 0 {
		invokers = newInvokerList(handlers)
	} else {
		if method == http.MethodOptions {
			invokers = s.invokers.options
		} else if path == faviconPath {
			invokers = s.invokers.favicon
		} else {
			invokers = s.invokers.notfound
		}
	}
	var resp Responder
	if s.BeginHandler != nil && !reservedPaths[path] {
		resp = s.BeginHandler.HandleRequest(ctx, req, invokers.Invoke)
	} else {
		resp = invokers.Invoke(ctx, req)
	}
	if resp == nil {
		resp = handleNotImplemented(ctx, req, nil)
	}
	resp.Respond(ctx, rw)
}

func (s *Server) wrapResponseWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	for k, v := range s.Header {
		rw.Header()[k] = v
	}
	w := io.NewResponseWriter(rw)
	if !s.CompressionEnabled {
		return w
	}

	enc := req.Header.Get("Accept-Encoding")
	if gox.IndexOfString(acceptEncodings, enc) < 0 {
		return w
	}
	cw, err := io.NewCompressResponseWriter(w, enc)
	if err != nil {
		log.Errorf("NewCompressResponseWriter: %v", err)
		return w
	}
	return cw
}

func (s *Server) initSession(rw http.ResponseWriter, req *http.Request) string {
	var sid string
	// Read cookie
	for _, c := range req.Cookies() {
		if c.Name == s.sessionName {
			sid = c.Value
			break
		}
	}

	// Read Header
	if sid == "" {
		lcName := strings.ToLower(s.sessionName)
		for k, vs := range req.Header {
			if strings.ToLower(k) == lcName {
				if len(vs) > 0 {
					sid = vs[0]
					break
				}
			}
		}
	}

	// Read url query
	if sid == "" {
		sid = req.URL.Query().Get(s.sessionName)
	}

	if sid == "" {
		sid = gox.UniqueID40()
	}

	cookie := &http.Cookie{
		Name:    s.sessionName,
		Value:   sid,
		Expires: time.Now().Add(s.sessionTTL),
		Path:    "/",
	}
	http.SetCookie(rw, cookie)
	// Write to Header in case cookie is disabled by some browsers
	rw.Header().Set(s.sessionName, sid)
	return sid
}

func (s *Server) setupContext(ctx context.Context, rw http.ResponseWriter, sid string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	ctx = withTemplate(ctx, s.templates)
	ctx = withResponseWriter(ctx, rw)
	ctx = withSessionID(ctx, sid)
	return ctx, cancel
}

func (s *Server) closeWriter(w http.ResponseWriter) {
	if cw, ok := w.(*io.CompressResponseWriter); ok {
		err := cw.Close()
		if err != nil {
			logger.Errorf("Close compressed response writer: %v", err)
		}
	}
}

func (s *Server) logHTTP(rw http.ResponseWriter, req *http.Request, startAt time.Time) {
	status := 0
	if w, ok := rw.(*io.ResponseWriter); ok {
		status = w.Status()
	}
	info := fmt.Sprintf("%s %s %s | %d %v",
		req.RemoteAddr,
		req.Method,
		req.RequestURI,
		status,
		time.Since(startAt))

	if status >= http.StatusBadRequest {
		if status != http.StatusUnauthorized {
			s.logger.Errorf("%s | %v | %v", info, req.Header, req.PostForm)
		} else {
			s.logger.Errorf("%s | %v", info, req.Header)
		}
	} else {
		s.logger.Info(info)
	}
}

func (s *Server) handleOptions(ctx context.Context, req *Request, next Invoker) Responder {
	path := req.NormalizedPath()
	var allowedMethods []string
	for routeMethod := range s.Router.methodToRoot {
		if handlers, _ := s.match(routeMethod, path); handlers != nil && handlers.Len() > 0 {
			allowedMethods = append(allowedMethods, routeMethod)
		}
	}

	if len(allowedMethods) > 0 {
		allowedMethods = append(allowedMethods, http.MethodOptions)
	}

	return ResponderFunc(func(ctx context.Context, rw http.ResponseWriter) {
		if len(allowedMethods) > 0 {
			val := []string{strings.Join(allowedMethods, ",")}
			rw.Header()["Allow"] = val
			rw.Header()["Access-Control-Allow-Methods"] = val
			rw.WriteHeader(http.StatusNoContent)
		} else {
			rw.WriteHeader(http.StatusNotFound)
		}
	})
}
