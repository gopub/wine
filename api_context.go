package wine

import (
	"html/template"
	"log"
	"net/http"

	"github.com/natande/gox"
)

//APIContext implements Content with DefaultContext
type APIContext struct {
	*DefaultContext
	userID   gox.ID
	handlers *HandlerChain
}

// Rebuild rebuilds the context
func (c *APIContext) Rebuild(
	rw http.ResponseWriter,
	req *http.Request,
	templates []*template.Template,
	handlers []Handler,
	maxMemory int64,
) {
	if c.DefaultContext == nil {
		c.DefaultContext = &DefaultContext{}
	}
	c.DefaultContext.Rebuild(rw, req, templates, handlers, maxMemory)
	c.handlers = NewHandlerChain(handlers)
}

// Next invokes the next handler
func (c *APIContext) Next() {
	if h := c.handlers.Next(); h != nil {
		h.HandleRequest(c)
	}
}

// SendResponse sends a response
func (c *APIContext) SendResponse(code int, msg string, data interface{}) {
	c.JSON(gox.M{"code": code, "data": data, "msg": msg})
}

// SendData sends a data response
func (c *APIContext) SendData(data interface{}) {
	c.SendResponse(0, "", data)
}

// SendError sends an error response
func (c *APIContext) SendError(err error) {
	var e *gox.Error
	if err == nil {
		e = gox.Success
	} else {
		e = gox.ParseError(err)
		if e == nil {
			e = gox.NewError(gox.EcodeServer, err.Error())
		}
	}
	c.SendResponse(e.Code(), e.Msg(), nil)
	if e.Code() != 0 {
		log.Println("[WINE] SendError:", e, c.Request(), c.Params())
	}
}

// SendMessage sends a message response
func (c *APIContext) SendMessage(code int, msg string) {
	c.SendResponse(code, msg, nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, msg, c.Request())
	}
}

// SetUserID sets current user id
func (c *APIContext) SetUserID(userID gox.ID) {
	c.userID = userID
	log.Println("[WINE] Set uid[", userID, "]", c.Request().URL)
}

// UserID return current user id
func (c *APIContext) UserID() gox.ID {
	return c.userID
}
