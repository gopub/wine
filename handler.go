package wine

type Handler interface {
	HandleRequest(Context)
}

type HandlerFunc func(Context)

func (this HandlerFunc) HandleRequest(c Context) {
	this(c)
}

type HandlerChain struct {
	index    int
	handlers []Handler
}

func NewHandlerChain(handlers []Handler) *HandlerChain {
	hc := &HandlerChain{}
	hc.handlers = handlers
	return hc
}

func (this *HandlerChain) Next() Handler {
	if this.index >= len(this.handlers) {
		return nil
	}

	index := this.index
	this.index += 1
	return this.handlers[index]
}
