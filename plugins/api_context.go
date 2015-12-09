package plugins

import (
	"github.com/justintan/gox"
	"github.com/justintan/gox/api"
	"github.com/justintan/wine"
	"html/template"
	"net/http"
)

var _, _ = api.Context(nil).(*APIContext)

//type Context wine.DefaultContext
type APIContext struct {
	*wine.DefaultContext
	userId   gox.Id
	req      *api.Request
	handlers *wine.HandlerChain
}

func NewAPIContext(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []wine.Handler) wine.Context {
	ctx := wine.NewDefaultContext(rw, req, templates, handlers).(*wine.DefaultContext)
	c := &APIContext{}
	c.DefaultContext = ctx
	c.handlers = wine.NewHandlerChain(handlers)
	h := gox.M{}
	for k, v := range c.Header() {
		h[k] = v
	}
	c.req = api.NewRequest("", h, c.Params())
	return c
}

func (c *APIContext) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

func (c *APIContext) Request() *api.Request {
	return c.req
}

func (c *APIContext) SendResponse(resp *api.Response) {
	c.JSON(resp)
}

func (c *APIContext) SendData(data interface{}) {
	c.SendResponse(api.NewResponse(api.OK, "", data))
}

func (c *APIContext) SendCode(code api.Code) {
	c.SendResponse(api.NewResponse(code, code.String(), nil))
	if code != api.OK {
		gox.LError(code, c.HTTPRequest())
	}
}

func (c *APIContext) SendMsg(code api.Code, msg string) {
	c.SendResponse(api.NewResponse(code, msg, nil))
	if code != api.OK {
		gox.LError(code, msg, c.HTTPRequest())
	}
}

func (c *APIContext) SetUserId(userId gox.Id) {
	c.userId = userId
	gox.LInfo("set uid[", userId, "]", c.HTTPRequest().URL)
}

func (c *APIContext) UserId() gox.Id {
	return c.userId
}
