package wine

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gopub/environ"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/resource"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/internal/template"
)

const (
	faviconPath = "favicon.ico"

	defaultReqMaxMem  = int(8 * types.MB)
	defaultSessionTTL = 30 * time.Minute
	minSessionTTL     = 5 * time.Minute
	defaultTimeout    = 10 * time.Second
)

var reservedPaths = map[string]bool{
	datePath:     true,
	uptimePath:   true,
	versionPath:  true,
	endpointPath: true,
	echoPath:     true,
	faviconPath:  true,
}

// Server implements web server
type Server struct {
	*Router
	*template.Manager
	server      *http.Server
	sessionTTL  time.Duration
	sessionName string

	maxReqMem          types.ByteUnit
	Header             http.Header
	Timeout            time.Duration
	PreHandler         Handler
	CompressionEnabled bool
	Recovery           bool
	ResultLogger       func(req *Request, result *Result, cost time.Duration)
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
		maxReqMem:          types.ByteUnit(environ.SizeInBytes("wine.max_memory", defaultReqMaxMem)),
		Router:             NewRouter(),
		Manager:            template.NewManager(),
		Header:             header,
		Timeout:            environ.Duration("wine.timeout", defaultTimeout),
		CompressionEnabled: environ.Bool("wine.compression", true),
		Recovery:           environ.Bool("wine.recovery", true),
		ResultLogger:       logResult,
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
func (s *Server) ServeHTTP(rw http.ResponseWriter, httpReq *http.Request) {
	startAt := time.Now()
	if s.Recovery {
		defer func() {
			if e := recover(); e != nil {
				logger.Errorf("%v: %+v\n", httpReq, e)
				logger.Errorf("\n%s\n", string(debug.Stack()))
			}
		}()
	}
	rw = s.wrapResponseWriter(rw, httpReq)
	defer s.closeWriter(rw)
	sid := s.initSession(rw, httpReq)
	ctx, cancel := s.setupContext(httpReq.Context())
	defer cancel()

	req, err := parseRequest(httpReq, s.maxReqMem)
	if err != nil {
		resp := Text(http.StatusBadRequest, fmt.Sprintf("Parse request: %v", err))
		resp.Respond(ctx, rw)
		s.logResult(&Request{request: httpReq}, rw, startAt)
		return
	}
	req.sid = sid
	req.params[s.sessionName] = sid
	ctx = s.withRequestParams(ctx, req.params)
	s.serve(ctx, req, rw)
	s.logResult(&Request{request: httpReq}, rw, startAt)
}

func (s *Server) serve(ctx context.Context, req *Request, rw http.ResponseWriter) {
	path := req.NormalizedPath()
	method := strings.ToUpper(req.Request().Method)
	r, params := s.Match(method, path)
	for k, v := range params {
		req.params[k] = v
	}
	var h Handler
	switch {
	case r != nil:
		if m := r.Model(); m != nil {
			if err := req.bind(m); err != nil {
				Error(err).Respond(ctx, rw)
				return
			}
			ctx = log.BuildContext(ctx, log.FromContext(ctx).With("model", m))
		}
		h = (*handlerElem)(r.FirstHandler())
	case method == http.MethodOptions:
		h = HandlerFunc(s.handleOptions)
	case path == faviconPath:
		h = HandleResponder(respond.Bytes(http.StatusOK, resource.Favicon))
	default:
		h = HandleResponder(Status(http.StatusNotFound))
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

func (s *Server) setupContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	ctx = withTemplateManager(ctx, s.Manager)
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

func (s *Server) handleOptions(_ context.Context, req *Request) Responder {
	methods := s.MatchScopes(req.NormalizedPath())
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

func (s *Server) logResult(req *Request, rw http.ResponseWriter, startAt time.Time) {
	res := new(Result)
	if getStatus, ok := rw.(interface{ Status() int }); ok {
		res.Status = getStatus.Status()
	}
	if getBody, ok := rw.(interface{ Body() []byte }); ok {
		res.Body = getBody.Body()
	}
	if s.ResultLogger != nil {
		s.ResultLogger(req, res, time.Since(startAt))
	}
}

func logResult(req *Request, res *Result, cost time.Duration) {
	httpReq := req.Request()
	if reservedPaths[httpReq.URL.Path[1:]] && res.Status < http.StatusBadRequest {
		return
	}
	info := fmt.Sprintf("%s %s %s | %d %v",
		httpReq.RemoteAddr,
		httpReq.Method,
		httpReq.RequestURI,
		res.Status,
		cost)
	if res.Status >= http.StatusBadRequest {
		ua := req.Header("User-Agent")
		if len(req.params) > 0 {
			info = fmt.Sprintf("%s | %s | %v", info, ua, JSONString(req.params))
		} else if len(httpReq.PostForm) > 0 {
			info = fmt.Sprintf("%s | %s | %v", info, ua, JSONString(httpReq.PostForm))
		} else {
			info = fmt.Sprintf("%s | %s", info, ua)
		}
		if req.uid > 0 {
			info = fmt.Sprintf("%s | user=%d", info, req.uid)
		}
		if len(res.Body) > 0 {
			logger.Errorf("%s | %s", info, res.Body)
		} else {
			logger.Errorf(info)
		}
	} else {
		if req.uid > 0 {
			logger.Infof("%s | uid=%d", info, req.uid)
		} else {
			logger.Info(info)
		}
	}
}
