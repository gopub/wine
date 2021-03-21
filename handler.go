package wine

import (
	"container/list"
	"context"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"

	"github.com/gopub/wine/urlutil"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req *Request) Responder
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req *Request) Responder

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req *Request) Responder {
	return h(ctx, req)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

type handlerElem list.Element

func (h *handlerElem) Next() *handlerElem {
	return (*handlerElem)((*list.Element)(h).Next())
}

func (h *handlerElem) HandleRequest(ctx context.Context, req *Request) Responder {
	return h.Value.(Handler).HandleRequest(withNextHandler(ctx, h.Next()), req)
}

func HTTPHandler(h http.Handler) Handler {
	return HandlerFunc(func(ctx context.Context, req *Request) Responder {
		return Handle(req.request, h)
	})
}

func HTTPHandlerFunc(h http.Handler) HandlerFunc {
	return func(ctx context.Context, req *Request) Responder {
		return Handle(req.request, h)
	}
}

func Prefix(prefix string, h Handler) Handler {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return h
	}

	return HandlerFunc(func(ctx context.Context, req *Request) Responder {
		u := req.request.URL
		r2 := new(http.Request)
		*r2 = *req.request
		r2.URL = new(url.URL)
		*r2.URL = *u
		r2.URL.Path = urlutil.Join(prefix, u.Path)
		r2.URL.RawPath = urlutil.Join(prefix, u.RawPath)
		req.request = r2
		return h.HandleRequest(ctx, req)
	})
}
