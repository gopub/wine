package wine

// Handler defines interface for interceptor
type Handler interface {
	HandleRequest(*Context)
}

// HandlerFunc converts function into Handler
type HandlerFunc func(*Context)

// HandleRequest is an interface method required by Handler
func (h HandlerFunc) HandleRequest(c *Context) {
	h(c)
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
