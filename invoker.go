package wine

import (
	"container/list"
	"context"
)

// Invoker defines the function to be called in order to pass on the request
type Invoker func(ctx context.Context, req *Request) Responder

type invokerList struct {
	handlers *list.List
	current  *list.Element
}

func toInvokerList(handlers ...Handler) *invokerList {
	hl := list.New()
	for _, h := range handlers {
		hl.PushBack(h)
	}
	return newInvokerList(hl)
}

func newInvokerList(hl *list.List) *invokerList {
	l := &invokerList{
		handlers: hl,
		current:  hl.Front(),
	}
	return l
}

func (l *invokerList) Invoke(ctx context.Context, req *Request) Responder {
	if l.current == nil {
		return nil
	}
	h := l.current.Value.(Handler)
	l.current = l.current.Next()
	ctx = withNext(ctx, l.Invoke)
	return h.HandleRequest(ctx, req)
}
