package plugins

import (
	"github.com/justintan/gox"
	"github.com/justintan/wine"
	"net/http"
)

//type Context wine.DefaultContext
type APIContext struct {
	*wine.DefaultContext
}

func NewAPIContext(rw http.ResponseWriter, req *http.Request, handlers []wine.Handler) wine.Context {
	ctx := wine.NewDefaultContext(rw, req, handlers).(*wine.DefaultContext)
	c := &APIContext{}
	c.DefaultContext = ctx
	return c
}

func (this *APIContext) Next() {
	if h := this.HandlerChain().Next(); h != nil {
		h(this)
	}
}

func (this *APIContext) SendData(data gox.M) {
	this.SendJSON(gox.M{"code": gox.OK, "data": data})
}

func (this *APIContext) SendCode(code gox.Code) {
	this.SendJSON(gox.M{"code": code, "msg": gox.MsgForCode(code)})
}

func (this *APIContext) User() *gox.User {
	user, _ := this.Get("user").(*gox.User)
	return user
}

func (this *APIContext) SetUser(user *gox.User) {
	this.Set("user", user)
}
