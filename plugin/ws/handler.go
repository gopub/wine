package ws

import (
	"container/list"
	"context"
	"reflect"
	"runtime"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req interface{}) (interface{}, error)
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req interface{}) (interface{}, error)

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return h(ctx, req)
}

func (h HandlerFunc) String() string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

type handlerElem list.Element

func (h *handlerElem) Next() *handlerElem {
	return (*handlerElem)((*list.Element)(h).Next())
}

func (h *handlerElem) HandleRequest(ctx context.Context, req interface{}) (interface{}, error) {
	return h.Value.(Handler).HandleRequest(withNextHandler(ctx, h.Next()), req)
}
