package main

import (
	"github.com/justintan/wine"
	"html/template"
	"net/http"
	"time"
)

type MyContext struct {
	*wine.DefaultContext
	handlers *wine.HandlerChain
}

func (c *MyContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []wine.Handler) {
	if c.DefaultContext == nil {
		c.DefaultContext = &wine.DefaultContext{}
	}
	c.DefaultContext.Rebuild(rw, req, templates, handlers)
	c.handlers = wine.NewHandlerChain(handlers)
}

func (c *MyContext) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

func (c *MyContext) SendResponse(code int, msg string, data interface{}) {
	c.JSON(map[string]interface{}{"code": code, "data": data, "msg": msg})
}

func main() {
	s := wine.Default()
	s.RegisterContext(&MyContext{})
	s.Get("time", func(c wine.Context) {
		ctx := c.(*MyContext)
		ctx.SendResponse(0, "", time.Now().Unix())
	})
	s.Run(":8000")
}
