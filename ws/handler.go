package ws

import (
	"container/list"
	"context"
	"reflect"
	"runtime"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req *Request) *Response
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req *Request) *Response

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req *Request) *Response {
	return h(ctx, req)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

type handlerElem list.Element

func (h *handlerElem) Next() *handlerElem {
	return (*handlerElem)((*list.Element)(h).Next())
}

func (h *handlerElem) HandleRequest(ctx context.Context, req *Request) *Response {
	return h.Value.(Handler).HandleRequest(withNextHandler(ctx, h.Next()), req)
}

func linkHandlers(handlers []Handler) *list.List {
	hl := list.New()
	for _, h := range handlers {
		hl.PushBack(h)
	}
	return hl
}

func linkHandlerFuncs(funcs []HandlerFunc) *list.List {
	hl := list.New()
	for _, h := range funcs {
		hl.PushBack(h)
	}
	return hl
}
