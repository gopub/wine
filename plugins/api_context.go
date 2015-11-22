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

func (this *APIContext) Next() {
	if h := this.handlers.Next(); h != nil {
		h.HandleRequest(this)
	}
}

func (this *APIContext) Request() *api.Request {
	return this.req
}

func (this *APIContext) SendResponse(resp *api.Response) {
	this.JSON(resp)
}

func (this *APIContext) SendData(data interface{}) {
	this.SendResponse(api.NewResponse(api.OK, "", data))
}

func (this *APIContext) SendCode(code api.Code) {
	this.SendResponse(api.NewResponse(code, code.String(), nil))
	if code != api.OK {
		gox.LDebug(this.HTTPRequest(), code, code.String())
	}
}

func (this *APIContext) SendMsg(code api.Code, msg string) {
	this.SendResponse(api.NewResponse(code, msg, nil))
	if code != api.OK {
		gox.LDebug(this.HTTPRequest(), code, msg)
	}
}

func (this *APIContext) SetUserId(userId gox.Id) {
	this.userId = userId
	gox.LDebug(this.HTTPRequest(), "set user_id:", userId)
}

func (this *APIContext) UserId() gox.Id {
	return this.userId
}
