package wine

import (
	"github.com/natande/gox"
	"net/http"
)

// Context is a default implementation of Context interface
type Context struct {
	Responder
	handlers  *handlerChain
	req       *http.Request
	reqParams gox.M
	keyValues gox.M
}

// Set sets key:value
func (c *Context) Set(key string, value interface{}) {
	c.keyValues[key] = value
}

// Get returns value for key
func (c *Context) Get(key string) interface{} {
	return c.keyValues[key]
}

// Next calls the next handler
func (c *Context) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

// Request returns request
func (c *Context) Request() *http.Request {
	return c.req
}

// Params returns request's parameters including queries, body
func (c *Context) Params() gox.M {
	return c.reqParams
}