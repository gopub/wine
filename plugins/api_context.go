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
	userId gox.Id
	req    *api.Request
}

func NewAPIContext(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []wine.Handler) wine.Context {
	ctx := wine.NewDefaultContext(rw, req, templates, handlers).(*wine.DefaultContext)
	c := &APIContext{}
	c.DefaultContext = ctx
	h := gox.M{}
	for k, v := range c.RequestHeader() {
		h[k] = v
	}
	c.req = api.NewRequest("", h, c.RequestParams())
	return c
}

func (this *APIContext) Next() {
	if h := this.HandlerChain().Next(); h != nil {
		h(this)
	}
}

func (this *APIContext) Request() *api.Request {
	return this.req
}

func (this *APIContext) SendResponse(resp *api.Response) {
	this.SendJSON(resp)
}

func (this *APIContext) SendData(data interface{}) {
	this.SendResponse(api.NewResponse(api.OK, "", data))
}

func (this *APIContext) SendCode(code api.Code) {
	this.SendResponse(api.NewResponse(code, code.String(), nil))
}

func (this *APIContext) SendMsg(code api.Code, msg string) {
	this.SendResponse(api.NewResponse(code, msg, nil))
}

func (this *APIContext) SetUserId(userId gox.Id) {
	this.userId = userId
}

func (this *APIContext) UserId() gox.Id {
	return this.userId
}
