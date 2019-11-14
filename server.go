package wine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gopub/wine/internal/request"

	"github.com/gopub/gox"

	"github.com/gopub/log"
	pathutil "github.com/gopub/wine/internal/path"
	"github.com/gopub/wine/internal/template"
)

var acceptEncodings = [2]string{"gzip", "defalte"}
var ShortHandlerNameFlag = true

const (
	endpointPath = "_endpoints"
	faviconPath  = "favicon.ico"
)

var SessionName = "wsessionid"
var SessionTTL = 30 * time.Minute

const minSessionTTL = 5 * time.Minute

// Server implements web server
type Server struct {
	*Router
	*TemplateManager
	server *http.Server

	Header       http.Header
	Timeout      time.Duration //Timeout for each request, default value is 20s
	ParamsParser ParamsParser
	BeginHandler Handler

	defaultHandler struct {
		favicon  *handlerList
		notfound *handlerList
		options  *handlerList
	}

	logger log.Logger

	reservedPaths map[string]bool
}

// NewServer returns a server
func NewServer() *Server {
	logger := log.GetLogger("Wine")
	logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)
	header := make(http.Header, 1)
	header.Set("Server", "Wine")

	s := &Server{
		Router:          NewRouter(),
		TemplateManager: NewTemplateManager(),
		Header:          header,
		Timeout:         10 * time.Second,
		ParamsParser:    request.NewParamsParser(8 * gox.MB),
		logger:          logger,
	}

	s.defaultHandler.favicon = newHandlerList([]Handler{HandlerFunc(handleFavIcon)})
	s.defaultHandler.notfound = newHandlerList([]Handler{HandlerFunc(handleNotFound)})
	s.defaultHandler.options = newHandlerList([]Handler{HandlerFunc(s.handleOptions)})
	s.reservedPaths = map[string]bool{
		endpointPath: true,
		faviconPath:  true,
	}
	s.AddTemplateFuncMap(template.FuncMap)
	return s
}

// Run starts server
func (s *Server) Run(addr string) {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	logger.Info("Running at", addr, "...")
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	if err != nil {
		logger.Panicf("ListenAndServe: %v", err)
	}
}

// RunTLS starts server with tls
func (s *Server) RunTLS(addr, certFile, keyFile string) error {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	s.Router.Print()
	logger.Info("Running at", addr, "...")
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServeTLS(certFile, keyFile)
	if err != nil {
		s.server = nil
	}
	return err
}

// Shutdown stops server
func (s *Server) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func (s *Server) logHTTP(rw http.ResponseWriter, req *http.Request, startAt time.Time) {
	var statGetter statusGetter
	if cw, ok := rw.(*CompressedResponseWriter); ok {
		cw.Close()
		statGetter = cw.ResponseWriter.(statusGetter)
	}

	if statGetter == nil {
		statGetter = rw.(statusGetter)
	}

	info := fmt.Sprintf("%s %s %s | return %d in %v",
		req.RemoteAddr,
		req.Method,
		req.RequestURI,
		statGetter.Status(),
		time.Since(startAt))

	if statGetter.Status() >= http.StatusBadRequest {
		if statGetter.Status() != http.StatusUnauthorized {
			s.logger.Errorf("%s Header: %v form: %v", info, req.Header, req.PostForm)
		} else {
			s.logger.Errorf("%s Header: %v", info, req.Header)
		}
	} else {
		s.logger.Info(info)
	}
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	sid := s.setupSession(rw, req)
	startAt := time.Now()
	if logger.Level() > log.DebugLevel {
		defer func() {
			if e := recover(); e != nil {
				logger.Error(e, req)
			}
		}()
	}

	// Add compression to responseWriter
	rw = createCompressedWriter(NewResponseWriter(rw), req)
	defer s.logHTTP(rw, req, startAt)

	path := pathutil.NormalizedRequestPath(req)
	method := strings.ToUpper(req.Method)
	handlers, pathParams := s.Match(method, path)

	if handlers.Empty() {
		if method == http.MethodOptions {
			handlers = s.defaultHandler.options
		} else if path == faviconPath {
			handlers = s.defaultHandler.favicon
		} else {
			handlers = s.defaultHandler.notfound
		}
	}

	ctx, cancel := context.WithTimeout(req.Context(), s.Timeout)
	defer cancel()

	parsedReq, err := NewRequest(req, s.ParamsParser)
	if err != nil {
		logger.Errorf("Parse failed: %v", err)
		resp := Text(http.StatusBadRequest, fmt.Sprintf("Parse request failed: %v", err))
		resp.Respond(ctx, rw)
		return
	}
	parsedReq.params.AddMapObj(pathParams)
	parsedReq.params[SessionName] = sid

	for k, v := range s.Header {
		rw.Header()[k] = v
	}

	ctx = withTemplate(ctx, s.templates)
	ctx = withResponseWriter(ctx, rw)
	var resp Responsible
	if s.BeginHandler != nil && !s.reservedPaths[path] {
		resp = s.BeginHandler.HandleRequest(ctx, parsedReq, handlers.Head().Invoke)
	} else {
		resp = handlers.Head().Invoke(ctx, parsedReq)
	}

	if resp == nil {
		resp = handleNotImplemented(ctx, parsedReq, nil)
	}
	resp.Respond(ctx, rw)
}

func createCompressedWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	// Add compression to responseWriter
	ae := req.Header.Get("Accept-Encoding")
	for _, enc := range acceptEncodings {
		if strings.Contains(ae, enc) {
			rw.Header().Set("Content-Encoding", enc)
			cw, err := NewCompressedResponseWriter(rw, enc)
			if err != nil {
				log.Errorf("NewCompressedResponseWriter failed: %v", err)
			}
			return cw
		}
	}

	return rw
}

func (s *Server) handleOptions(ctx context.Context, req *Request, next Invoker) Responsible {
	path := pathutil.NormalizedRequestPath(req.request)
	var allowedMethods []string
	for routeMethod := range s.Router.methodTrees {
		if handlers, _ := s.Match(routeMethod, path); !handlers.Empty() {
			allowedMethods = append(allowedMethods, routeMethod)
		}
	}

	if len(allowedMethods) > 0 {
		allowedMethods = append(allowedMethods, http.MethodOptions)
	}

	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
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

func (s *Server) setupSession(rw http.ResponseWriter, req *http.Request) string {
	if SessionName == "" {
		return ""
	}

	var sid string
	// Read cookie
	for _, c := range req.Cookies() {
		if c.Name == SessionName {
			sid = c.Value
			break
		}
	}

	// Read Header
	if sid == "" {
		lcName := strings.ToLower(SessionName)
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
		sid = req.URL.Query().Get(SessionName)
	}

	if sid == "" {
		sid = gox.UniqueID()
	}

	var expires time.Time
	if SessionTTL < minSessionTTL {
		expires = time.Now().Add(minSessionTTL)
	} else {
		expires = time.Now().Add(SessionTTL)
	}
	cookie := &http.Cookie{
		Name:    SessionName,
		Value:   sid,
		Expires: expires,
		Path:    "/",
	}
	http.SetCookie(rw, cookie)
	// Write to Header in case cookie is disabled by some browsers
	rw.Header().Set(SessionName, sid)
	return sid
}
