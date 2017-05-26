package wine

type Handler interface {
	HandleRequest(Context)
}

type HandlerFunc func(Context)

func (h HandlerFunc) HandleRequest(c Context) {
	h(c)
}

// HandlerChain : A chain of handlers
type HandlerChain struct {
	index    int
	handlers []Handler
}

// NewHandlerChain : Create handler chain
func NewHandlerChain(handlers []Handler) *HandlerChain {
	hc := &HandlerChain{}
	hc.handlers = handlers
	return hc
}

// Next : Get next handler
func (h *HandlerChain) Next() Handler {
	if h.index >= len(h.handlers) {
		return nil
	}

	index := h.index
	h.index++
	return h.handlers[index]
}
