package wine

import (
	"context"
	"sync"
)

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(ctx context.Context, req Request, next Invoker) Responsible
}

// HandlerFunc converts function into Handler
type HandlerFunc func(ctx context.Context, req Request, next Invoker) Responsible

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req Request, next Invoker) Responsible {
	return h(ctx, req, next)
}

type Invoker func(ctx context.Context, req Request) Responsible

type handlerElement struct {
	handler Handler
	next    *handlerElement
}

func (h *handlerElement) Invoke(ctx context.Context, req Request) Responsible {
	return h.handler.HandleRequest(ctx, req, h.next.Invoke)
}

type handlerList struct {
	head *handlerElement
	tail *handlerElement
	mu   sync.Mutex
}

func (l *handlerList) Empty() bool {
	if l == nil {
		return true
	}
	return l.head == nil
}

func (l *handlerList) Head() *handlerElement {
	return l.head
}

func (l *handlerList) Tail() *handlerElement {
	return l.tail
}

func (l *handlerList) PushBack(v Handler) {
	l.mu.Lock()
	e := &handlerElement{handler: v, next: nil}
	if l.tail == nil {
		l.head = e
		l.tail = e
	} else {
		l.tail.next = e
		l.tail = l.tail.next
	}
	l.mu.Unlock()
}

func newHandlerList(handlers []Handler) *handlerList {
	l := &handlerList{}
	for _, h := range handlers {
		l.PushBack(h)
	}
	return l
}
