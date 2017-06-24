package wine

import (
	"html/template"
	"log"
	"net/http"

	"github.com/natande/gox"
)

var _ Responder = (*APIResponder)(nil)

//APIResponder is designed for api request/response
type APIResponder struct {
	*DefaultResponder
	loginID gox.ID
}

// Reset resets the Responder
func (ar *APIResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	if ar.DefaultResponder == nil {
		ar.DefaultResponder = &DefaultResponder{}
	}
	ar.DefaultResponder.Reset(req, rw, tmpls)
}

// SendResponse sends a response
func (ar *APIResponder) SendResponse(code int, msg string, data interface{}) {
	ar.JSON(gox.M{"code": code, "data": data, "msg": msg})
}

// SendData sends a data response
func (ar *APIResponder) SendData(data interface{}) {
	ar.SendResponse(0, "", data)
}

// SendError sends an error response
func (ar *APIResponder) SendError(err error) {
	var e *gox.Error
	if err == nil {
		e = gox.ErrSuccess
	} else {
		e = gox.ParseError(err)
		if e == nil {
			e = gox.NewError(gox.EcodeServer, erar.Error())
		}
	}
	ar.SendResponse(e.Code(), e.Msg(), nil)
	if e.Code() != 0 {
		log.Println("[WINE] SendError:", e, ar.req)
	}
}

// SendMessage sends a message response
func (ar *APIResponder) SendMessage(code int, msg string) {
	ar.SendResponse(code, msg, nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, msg, ar.req)
	}
}

// SetLoginID sets current login user id
func (ar *APIResponder) SetLoginID(loginID gox.ID) {
	ar.loginID = loginID
	log.Println("[WINE] Set user id[", loginID, "]", ar.req.URL)
}

// LoginID return current login user id
func (ar *APIResponder) LoginID() gox.ID {
	return ar.loginID
}
