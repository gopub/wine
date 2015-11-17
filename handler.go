package wine

type Handler interface {
	HandleRequest(Context)
}

type HandlerFunc func(Context)

func (this HandlerFunc) HandleRequest(c Context) {
	this(c)
}

type handlerChain struct {
	index    int
	handlers []Handler
}

func NewHandlerChain(handlers []Handler) *handlerChain {
	hc := &handlerChain{}
	hc.handlers = handlers
	return hc
}

func (this *handlerChain) Next() Handler {
	if this.index >= len(this.handlers) {
		return nil
	}

	index := this.index
	this.index += 1
	return this.handlers[index]
}
