package wine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gopub/wine/internal/resource"

	"github.com/gopub/environ"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/internal/template"
)

const (
	sysDatePath  = "_sys/date"
	sysUptime    = "_sys/uptime"
	sysVersion   = "_sys/version"
	endpointPath = "_debug/endpoints"
	echoPath     = "_debug/echo"
	faviconPath  = "favicon.ico"
	version      = "v1.22.21"
)

var serverUpAt = time.Now()

var reservedPaths = map[string]bool{
	sysDatePath:  true,
	sysUptime:    true,
	sysVersion:   true,
	endpointPath: true,
	faviconPath:  true,
	echoPath:     true,
}

const (
	defaultReqMaxMem  = int(8 * types.MB)
	defaultSessionTTL = 30 * time.Minute
	minSessionTTL     = 5 * time.Minute
	defaultTimeout    = 10 * time.Second
)

// Server implements web server
type Server struct {
	*Router
	*template.Manager
	server      *http.Server
	sessionTTL  time.Duration
	sessionName string

	maxRequestMemory   types.ByteUnit
	Header             http.Header
	Timeout            time.Duration
	PreHandler         Handler
	CompressionEnabled bool
	Recovery           bool
}

// NewServer returns a server
func NewServer() *Server {
	logger := log.GetLogger("Wine")
	logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)
	header := make(http.Header, 1)
	header.Set("Server", "Wine")

	s := &Server{
		sessionName:        environ.String("wine.session.name", "wsessionid"),
		sessionTTL:         environ.Duration("wine.session.ttl", defaultSessionTTL),
		maxRequestMemory:   types.ByteUnit(environ.SizeInBytes("wine.max_memory", defaultReqMaxMem)),
		Router:             NewRouter(),
		Manager:            template.NewManager(),
		Header:             header,
		Timeout:            environ.Duration("wine.timeout", defaultTimeout),
		CompressionEnabled: environ.Bool("wine.compression", true),
		Recovery:           environ.Bool("wine.recovery", true),
	}

	if s.sessionTTL < minSessionTTL {
		s.sessionTTL = minSessionTTL
	}
	s.AddTemplateFuncMap(template.FuncMap)
	return s
}

// Run starts server
func (s *Server) Run(addr string) {
	if s.server != nil {
		logger.Fatalf("Server is running")
	}

	logger.Infof("Running at %s ...", addr)
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
		log.Panic("Server is running")
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
	defer s.logRequest(req, rw, time.Now())

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
	ctx = s.withRequestParams(ctx, parsedReq.params)
	s.serve(ctx, parsedReq, rw)
}

func (s *Server) serve(ctx context.Context, req *Request, rw http.ResponseWriter) {
	path := req.NormalizedPath()
	method := strings.ToUpper(req.Request().Method)
	var h Handler
	hl, params := s.match(method, path)
	for k, v := range params {
		req.params[k] = v
	}
	if hl != nil && hl.Len() > 0 {
		h = (*handlerElem)(hl.Front())
	} else {
		if method == http.MethodOptions {
			h = HandlerFunc(s.handleOptions)
		} else if path == faviconPath {
			h = HandleResponder(respond.Bytes(http.StatusOK, resource.Favicon))
		} else {
			h = HandleResponder(Status(http.StatusNotFound))
		}
	}
	var resp Responder
	if s.PreHandler != nil && !reservedPaths[path] {
		resp = s.PreHandler.HandleRequest(withNextHandler(ctx, h), req)
	} else {
		resp = h.HandleRequest(ctx, req)
	}
	if resp == nil {
		resp = Status(http.StatusNotImplemented)
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

	encodings := strings.Split(req.Header.Get("Accept-Encoding"), ",")
	for _, enc := range encodings {
		enc = strings.TrimSpace(enc)
		cw, err := io.NewCompressResponseWriter(w, enc)
		if err == nil {
			return cw
		}
	}
	log.Warnf("Unsupported encodings: %v", encodings)
	return w
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
		sid = NewUUID()
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
	ctx = withTemplateManager(ctx, s.Manager)
	ctx = withSessionID(ctx, sid)
	return ctx, cancel
}

func (s *Server) withRequestParams(ctx context.Context, params types.M) context.Context {
	if loc, _ := types.NewPointFromString(params.String("coordinate")); loc != nil {
		ctx = WithCoordinate(ctx, loc)
	}
	if deviceID := params.String("device_id"); deviceID != "" {
		ctx = WithDeviceID(ctx, deviceID)
	}
	return ctx
}

func (s *Server) closeWriter(w http.ResponseWriter) {
	if cw, ok := w.(*io.CompressResponseWriter); ok {
		err := cw.Close()
		if err != nil {
			logger.Errorf("Close compressed response writer: %v", err)
		}
	}
}

func (s *Server) logRequest(req *http.Request, rw http.ResponseWriter, startAt time.Time) {
	status := 0
	if w, ok := rw.(interface{ Status() int }); ok {
		status = w.Status()
	}
	if reservedPaths[req.RequestURI[1:]] && status < http.StatusBadRequest {
		return
	}
	info := fmt.Sprintf("%s %s %s | %d %v",
		req.RemoteAddr,
		req.Method,
		req.RequestURI,
		status,
		time.Since(startAt))
	if status >= http.StatusBadRequest {
		cookie := req.Header.Get("Cookie")
		ua := req.Header.Get("User-Agent")
		if status != http.StatusUnauthorized {
			logger.Errorf("%s | Cookie:%s User-Agent:%s | %v", info, cookie, ua, req.PostForm)
		} else {
			logger.Errorf("%s | Cookie:%s User-Agent:%s", info, cookie, ua)
		}
	} else {
		logger.Info(info)
	}
}

func (s *Server) handleOptions(_ context.Context, req *Request) Responder {
	methods := s.matchMethods(req.NormalizedPath())
	if len(methods) > 0 {
		methods = append(methods, http.MethodOptions)
	}

	return respond.Func(func(ctx context.Context, rw http.ResponseWriter) {
		if len(methods) == 0 {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		joined := []string{strings.Join(methods, ",")}
		rw.Header()["Allow"] = joined
		rw.Header()["Access-Control-Allow-Methods"] = joined
		rw.WriteHeader(http.StatusNoContent)
	})
}
