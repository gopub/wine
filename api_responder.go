package wine

import (
	"html/template"
	"log"
	"net/http"

	"github.com/natande/gox"
)

var _ Responder = (*APIResponder)(nil)

//APIResponder implements Content with DefaultResponder
type APIResponder struct {
	*DefaultResponder
	userID gox.ID
}

// Reset resets the Responder
func (c *APIResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	if c.DefaultResponder == nil {
		c.DefaultResponder = &DefaultResponder{}
	}
	c.DefaultResponder.Reset(req, rw, tmpls)
}

// SendResponse sends a response
func (c *APIResponder) SendResponse(code int, msg string, data interface{}) {
	c.JSON(gox.M{"code": code, "data": data, "msg": msg})
}

// SendData sends a data response
func (c *APIResponder) SendData(data interface{}) {
	c.SendResponse(0, "", data)
}

// SendError sends an error response
func (c *APIResponder) SendError(err error) {
	var e *gox.Error
	if err == nil {
		e = gox.ErrSuccess
	} else {
		e = gox.ParseError(err)
		if e == nil {
			e = gox.NewError(gox.EcodeServer, err.Error())
		}
	}
	c.SendResponse(e.Code(), e.Msg(), nil)
	if e.Code() != 0 {
		log.Println("[WINE] SendError:", e, c.req)
	}
}

// SendMessage sends a message response
func (c *APIResponder) SendMessage(code int, msg string) {
	c.SendResponse(code, msg, nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, msg, c.req)
	}
}

// SetUserID sets current user id
func (c *APIResponder) SetUserID(userID gox.ID) {
	c.userID = userID
	log.Println("[WINE] Set uid[", userID, "]", c.req.URL)
}

// UserID return current user id
func (c *APIResponder) UserID() gox.ID {
	return c.userID
}
