package wine

import "context"

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(context.Context, Request, Responder) bool
}

// HandlerFunc converts function into Handler
type HandlerFunc func(context.Context, Request, Responder) bool

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(ctx context.Context, req Request, resp Responder) bool {
	return h(ctx, req, resp)
}

// HandlerChain : A chain of handlers
type handlerChain struct {
	index    int
	handlers []Handler
}

// NewHandlerChain : Create handler chain
func newHandlerChain(handlers []Handler) *handlerChain {
	hc := &handlerChain{}
	hc.handlers = handlers
	return hc
}

// Next : Get next handler
func (h *handlerChain) Next() Handler {
	if h.index >= len(h.handlers) {
		return nil
	}

	index := h.index
	h.index++
	return h.handlers[index]
}
