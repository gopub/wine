package wine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"mime"
	"net"
	"net/http"
	"path"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/gopub/conv"
	"github.com/gopub/environ"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"github.com/gopub/wine/ctxutil"
	"github.com/gopub/wine/httpvalue"
	"github.com/gopub/wine/internal/io"
	"github.com/gopub/wine/internal/resource"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/internal/template"
)

const (
	faviconPath = "favicon.ico"

	defaultReqMaxMem = int(8 * types.MB)
	defaultTimeout   = 10 * time.Second

	minAutoCompressionSize = 2048
)

var reservedPaths = map[string]bool{
	datePath:     true,
	uptimePath:   true,
	versionPath:  true,
	endpointPath: true,
	echoPath:     true,
	faviconPath:  true,
}

type Options struct {
	ReqFormMem      types.ByteUnit
	Timeout         time.Duration
	Recovery        bool
	AutoCompression bool
	LoggingReqModel bool
}

// Server implements web server
type Server struct {
	*Router
	*template.Manager
	server *http.Server

	addr string
	url  string

	Options
	ResultLogger    func(req *Request, result *Result, cost time.Duration)
	NotFoundHandler Handler
}

// NewServer returns a server
func NewServer(options *Options) *Server {
	logger := log.GetLogger("Wine")
	logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)
	header := make(http.Header, 1)
	header.Set("Server", "Wine")

	if options == nil {
		options = &Options{
			ReqFormMem:      types.ByteUnit(environ.SizeInBytes("wine.max_memory", defaultReqMaxMem)),
			Timeout:         environ.Duration("wine.timeout", defaultTimeout),
			Recovery:        environ.Bool("wine.recovery", true),
			AutoCompression: environ.Bool("wine.compression.auto", true),
			LoggingReqModel: environ.Bool("wine.logging.request.model", true),
		}
	}

	s := &Server{
		Router:       NewRouter(),
		Manager:      template.NewManager(),
		ResultLogger: logResult,
		Options:      *options,
	}

	s.AddTemplateFuncMap(template.FuncMap)
	return s
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) URL() string {
	return s.url
}

func (s *Server) assignAddr(addr string, tls bool) {
	s.addr = addr
	if strings.HasPrefix(addr, ":") {
		s.addr = "0.0.0.0" + addr
	} else {
		s.addr = addr
	}

	if addr == "" {
		s.url = ""
	} else {
		if tls {
			s.url = "https://" + s.addr
		} else {
			s.url = "http://" + s.addr
		}
	}
}

// Run starts server
func (s *Server) Run(addr string) {
	if s.server != nil {
		logger.Panicf("HTTP Server is running")
	}

	logger.Infof("HTTP server is running on %s", addr)
	s.server = &http.Server{Addr: addr, Handler: s}
	s.assignAddr(s.server.Addr, false)
	err := s.server.ListenAndServe()
	s.assignAddr("", false)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Infof("HTTP server was closed")
		} else {
			logger.Panicf("HTTP server was terminated: %v", err)
		}
	}
}

// RunTLS starts server with tls
func (s *Server) RunTLS(addr, certFile, keyFile string) {
	if s.server != nil {
		logger.Panic("HTTPS Server is running")
	}

	logger.Infof("HTTPS server is running on %s", addr)
	s.server = &http.Server{Addr: addr, Handler: s}
	s.assignAddr(s.server.Addr, true)
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	s.assignAddr("", false)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logger.Infof("HTTPS server was closed")
		} else {
			logger.Panicf("HTTPS server was terminated: %v", err)
		}
	}
}

// Shutdown shutdown server
func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server) Match(scope string, path string) (*Endpoint, map[string]string) {
	e, p := s.Router.Match(scope, path)
	return s.toEndpoint(e), p
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	startAt := time.Now()
	if s.Recovery {
		defer func() {
			if e := recover(); e != nil {
				logger.Errorf("%v: %+v\n", req, e)
				logger.Errorf("\n%s\n", string(debug.Stack()))
			}
		}()
	}
	rw = s.wrapResponseWriter(rw, req)
	ctx, cancel := s.initContext(req)
	defer cancel()

	wReq, err := parseRequest(req, s.ReqFormMem)
	if err != nil {
		defer s.closeWriter(rw)
		resp := Text(http.StatusBadRequest, fmt.Sprintf("Parse request: %v", err))
		resp.Respond(ctx, rw)
		s.logResult(&Request{request: req}, rw, startAt)
		return
	}
	s.serve(ctx, wReq, rw)
	s.logResult(wReq, rw, startAt)
}

func (s *Server) serve(ctx context.Context, req *Request, rw http.ResponseWriter) {
	np := req.NormalizedPath()
	method := req.Request().Method
	endpoint, params := s.Match(method, np)
	req.setPathParams(params)
	req.endpoint = endpoint
	s.Header().WriteTo(rw)
	var h Handler
	switch {
	case endpoint != nil:
		endpoint.Header().WriteTo(rw)
		req.sensitive = endpoint.Sensitive()
		if m := endpoint.Model(); m != nil {
			if err := req.bind(m); err != nil {
				Error(err).Respond(ctx, rw)
				return
			}
			if s.LoggingReqModel && !endpoint.Sensitive() {
				var logger *log.Logger
				if ls, ok := req.Model.(LogStringer); ok {
					logger = log.FromContext(ctx).With("model", ls.LogString())
				} else {
					logger = log.FromContext(ctx).With("model", conv.MustJSONString(req.Model))
				}
				ctx = log.BuildContext(ctx, logger)
			}
		}
		h = (*handlerElem)(endpoint.FirstHandler())
	case method == http.MethodOptions:
		h = HandlerFunc(s.handleOptions)
	case np == faviconPath:
		h = HandleResponder(respond.Bytes(http.StatusOK, resource.Favicon))
	default:
		if s.NotFoundHandler != nil {
			h = s.NotFoundHandler
		} else {
			h = HandleResponder(Status(http.StatusNotFound))
		}
	}

	resp := h.HandleRequest(ctx, req)
	if resp == nil {
		resp = Status(http.StatusNotImplemented)
	}
	rw = s.compressWriter(rw, req, resp)
	defer s.closeWriter(rw)
	resp.Respond(ctx, rw)
}

func (s *Server) compressWriter(w http.ResponseWriter, req *Request, responder Responder) http.ResponseWriter {
	if !s.AutoCompression {
		return w
	}

	respObj, ok := responder.(*respond.Response)
	if !ok {
		return w
	}

	if !httpvalue.IsMIMETextType(httpvalue.GetContentType(respObj.Header())) {
		return w
	}

	if respObj.ContentLength() < minAutoCompressionSize {
		return w
	}

	cw, err := CompressWriter(w, httpvalue.GetAcceptEncodings(req.request.Header)...)
	if err != nil {
		logger.Errorf("Cannot create compress writer: %w", err)
		return w
	}
	return cw
}

func (s *Server) wrapResponseWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	// contentType may be the final correct value. It can be overwritten
	contentType := mime.TypeByExtension(path.Ext(req.URL.Path))
	if contentType != "" && rw.Header().Get(httpvalue.ContentType) == "" {
		rw.Header().Set(httpvalue.ContentType, contentType)
	}
	s.Router.md.Header.WriteTo(rw)
	w := io.NewResponseWriter(rw)
	return w
}

func (s *Server) initContext(req *http.Request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(req.Context(), s.Timeout)
	ctx = ctxutil.WithTemplateManager(ctx, s.Manager)
	ctx = ctxutil.WithRequestHeader(ctx, req.Header)
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

func (s *Server) handleOptions(_ context.Context, req *Request) Responder {
	// TODO: how to handle preflight correctly?
	methods := s.MatchScopes(req.NormalizedPath())
	if len(methods) > 0 {
		methods = append(methods, http.MethodOptions)
	}
	return respond.Func(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header().Set(httpvalue.ACLAllowMethods, strings.Join(methods, ","))
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
		if reqID := req.Header(httpvalue.RequestID); reqID != "" {
			info = fmt.Sprintf("%s | %s | %s", info, ua, reqID)
		} else {
			info = fmt.Sprintf("%s | %s", info, ua)
		}
		if !req.sensitive {
			if len(req.Params()) > 0 {
				info = fmt.Sprintf("%s | %v", info, conv.MustJSONString(req.Params()))
			} else if len(httpReq.PostForm) > 0 {
				info = fmt.Sprintf("%s | %v", info, conv.MustJSONString(httpReq.PostForm))
			}
		}

		if req.uid > 0 {
			info = fmt.Sprintf("%s | user=%d", info, req.uid)
		}

		if len(res.Body) > 0 {
			if len(res.Body) < 2048 {
				logger.Errorf("%s | %s", info, res.Body)
			} else {
				logger.Errorf("%s | %d", info, len(res.Body))
			}
		} else {
			logger.Errorf(info)
		}
	} else {
		if req.uid > 0 {
			logger.Debugf("%s | uid=%d", info, req.uid)
		} else {
			logger.Debug(info)
		}
	}
}

type TestServer struct {
	*Server
	URL string
}

func NewTestServer(t *testing.T) *TestServer {
	s := NewServer(nil)
	t.Cleanup(func() {
		s.Shutdown()
	})
	return &TestServer{
		Server: s,
	}
}

func (s *TestServer) Run() string {
	for n := 0; n < 10; n++ {
		addr := fmt.Sprintf("localhost:%d", rand.Int()%1e4+1e4)
		conn, err := net.Dial("TCP", addr)
		if err != nil {
			go s.Server.Run(addr)
			s.URL = "http://" + addr
			break
		}
		conn.Close()
	}
	return s.URL
}

func (s *TestServer) RunTLS() string {
	// TODO:
	logger.Panic("Not implemented")
	return s.URL
}
