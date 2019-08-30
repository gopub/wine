package wine

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gopub/gox"
	"github.com/gopub/log"
)

const defaultMaxRequestMemory = 1 << 20
const defaultRequestTimeout = 20 * time.Second
const keyHTTPResponseWriter = "wine_http_response_writer"
const keyTemplates = "wine_templates"

var acceptEncodings = [2]string{"gzip", "defalte"}
var ShortHandlerNameFlag = true

// Server implements web server
type Server struct {
	*Router
	*TemplateManager

	Header           http.Header
	MaxRequestMemory int64         //max memory for request, default value is 8M
	RequestTimeout   time.Duration //timeout for each request, default value is 20s
	server           *http.Server

	h2conns *h2connCache

	RequestParser RequestParser

	faviconHandlerList  *handlerList
	notfoundHandlerList *handlerList
	optionsHandlerList  *handlerList

	logger log.Logger
}

// NewServer returns a server
func NewServer(config *Config) *Server {
	s := &Server{
		Router:           NewRouter(),
		TemplateManager:  NewTemplateManager(),
		Header:           make(http.Header),
		MaxRequestMemory: defaultMaxRequestMemory,
		RequestTimeout:   defaultRequestTimeout,
		h2conns:          newH2ConnCache(),
		RequestParser:    NewDefaultRequestParser(),

		logger: log.GetLogger("Wine"),
	}

	s.faviconHandlerList = newHandlerList([]Handler{HandlerFunc(handleFavIcon)})
	s.notfoundHandlerList = newHandlerList([]Handler{HandlerFunc(handleNotFound)})
	s.optionsHandlerList = newHandlerList([]Handler{HandlerFunc(s.handleOptions)})

	s.logger.SetFlags(logger.Flags() ^ log.Lfunction ^ log.Lshortfile)

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
	return s
}

// Run starts server
func (s *Server) Run(addr string) error {
	if s.server != nil {
		logger.Panic("Server is running")
	}

	s.Router.Print()
	logger.Info("Running at", addr, "...")
	s.server = &http.Server{Addr: addr, Handler: s}
	err := s.server.ListenAndServe()
	if err != nil {
		s.server = nil
	}
	return err
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
	if cw, ok := rw.(*compressedResponseWriter); ok {
		cw.Close()
		statGetter = cw.ResponseWriter.(statusGetter)
	}

	if statGetter == nil {
		statGetter = rw.(statusGetter)
	}

	info := fmt.Sprintf("%s %s %s %s | return %d in %v",
		req.RemoteAddr,
		req.UserAgent(),
		req.Method,
		req.RequestURI,
		statGetter.Status(),
		time.Since(startAt))

	if statGetter.Status() >= 400 {
		if statGetter.Status() != http.StatusUnauthorized {
			s.logger.Errorf("%s request: %v", info, req)
		} else {
			s.logger.Error(info)
		}
	} else {
		s.logger.Info(info)
	}
}

// ServeHTTP implements for http.Handler interface, which will handle each http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	startAt := time.Now()
	if logger.Level() > log.DebugLevel {
		defer func() {
			if e := recover(); e != nil {
				logger.Error(e, req)
			}
		}()
	}

	// Add compression to responseWriter
	rw = &responseWriterWrapper{ResponseWriter: rw}
	rw = wrapperCompressedWriter(rw, req)
	defer s.logHTTP(rw, req, startAt)

	path := getRequestPath(req)
	method := strings.ToUpper(req.Method)
	handlers, pathParams := s.Match(method, path)

	if handlers.Empty() {
		if method == http.MethodOptions {
			handlers = s.optionsHandlerList
		} else if path == "favicon.ico" {
			handlers = s.faviconHandlerList
		} else {
			handlers = s.notfoundHandlerList
		}
	}

	params, err := s.RequestParser.ParseHTTPRequest(req, s.MaxRequestMemory)
	if err != nil {
		logger.Errorf("ParseHTTPRequest failed: %v", err)
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
	ctx = context.WithValue(ctx, keyHTTPResponseWriter, rw)
	resp := handlers.Head().Invoke(ctx, parsedReq)
	if resp == nil {
		resp = handleNotImplemented(ctx, parsedReq, nil)
	}
	resp.Respond(ctx, rw)
}

func wrapperCompressedWriter(rw http.ResponseWriter, req *http.Request) http.ResponseWriter {
	// Add compression to responseWriter
	ae := req.Header.Get("Accept-Encoding")
	for _, enc := range acceptEncodings {
		if strings.Contains(ae, enc) {
			rw.Header().Set("Content-Encoding", enc)
			cw, err := newCompressedResponseWriter(rw, enc)
			if err != nil {
				log.Errorf("newCompressedResponseWriter failed: %v", err)
			}
			return cw
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

func (s *Server) handleOptions(ctx context.Context, req *Request, next Invoker) Responsible {
	path := getRequestPath(req.HTTPRequest)
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

func handleFavIcon(ctx context.Context, req *Request, next Invoker) Responsible {
	return ResponsibleFunc(func(ctx context.Context, rw http.ResponseWriter) {
		rw.Header()[ContentType] = []string{"image/x-icon"}
		rw.WriteHeader(http.StatusOK)
		if err := gox.WriteAll(rw, _faviconBytes); err != nil {
			log.ContextLogger(ctx).Error("cannot write bytes: %v", err)
		}
	})
}

func handleNotFound(ctx context.Context, req *Request, next Invoker) Responsible {
	return Text(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func handleNotImplemented(ctx context.Context, req *Request, next Invoker) Responsible {
	return Text(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(keyHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func GetTemplates(ctx context.Context) []*template.Template {
	v, _ := ctx.Value(keyTemplates).([]*template.Template)
	return v
}
