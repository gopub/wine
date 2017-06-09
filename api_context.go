package wine

import (
	"html/template"
	"net/http"

	"log"

	"github.com/justintan/gox/errors"
	"github.com/justintan/gox/types"
)

//APIContext implements Content with DefaultContext
type APIContext struct {
	*DefaultContext
	userID   types.ID
	handlers *HandlerChain
}

func (c *APIContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) {
	if c.DefaultContext == nil {
		c.DefaultContext = &DefaultContext{}
	}
	c.DefaultContext.Rebuild(rw, req, templates, handlers)
	c.handlers = NewHandlerChain(handlers)
}

func (c *APIContext) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

func (c *APIContext) SendResponse(code int, msg string, data interface{}) {
	c.JSON(types.M{"code": code, "data": data, "msg": msg})
}

func (c *APIContext) SendData(data interface{}) {
	c.SendResponse(0, "", data)
}

func (c *APIContext) SendCode(code int) {
	c.SendResponse(code, "", nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, c.HTTPRequest())
	}
}

func (c *APIContext) SendError(err error) {
	var e *errors.Error
	if err == nil {
		e = errors.Success
	} else {
		e = errors.ParseError(err)
		if e == nil {
			e = errors.NewError(errors.EcodeServer, err.Error())
		}
	}
	c.SendResponse(e.Code(), e.Msg(), nil)
	if e.Code() != 0 {
		log.Println("[WINE] Error:", e, c.HTTPRequest(), c.Params())
	}
}

func (c *APIContext) SendMessage(code int, msg string) {
	c.SendResponse(code, msg, nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, msg, c.HTTPRequest())
	}
}

func (c *APIContext) SetUserID(userID types.ID) {
	c.userID = userID
	log.Println("[WINE] Set uid[", userID, "]", c.HTTPRequest().URL)
}

func (c *APIContext) UserID() types.ID {
	return c.userID
}
