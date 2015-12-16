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
	userID   gox.ID
	req      *api.Request
	handlers *wine.HandlerChain
}

func (c *APIContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []wine.Handler) {
	if c.DefaultContext == nil {
		c.DefaultContext = &wine.DefaultContext{}
	}
	c.DefaultContext.Rebuild(rw, req, templates, handlers)
	c.handlers = wine.NewHandlerChain(handlers)
	h := gox.M{}
	for k, v := range c.Header() {
		h[k] = v
	}
	c.req = api.NewRequest("", h, c.Params())
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

func (c *APIContext) SetUserID(userID gox.ID) {
	c.userID = userID
	gox.LInfo("set uid[", userID, "]", c.HTTPRequest().URL)
}

func (c *APIContext) UserID() gox.ID {
	return c.userID
}
