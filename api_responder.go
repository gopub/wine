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
func (r *APIResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	if r.DefaultResponder == nil {
		r.DefaultResponder = &DefaultResponder{}
	}
	r.DefaultResponder.Reset(req, rw, tmpls)
}

// SendResponse sends a response
func (r *APIResponder) SendResponse(code int, msg string, data interface{}) {
	r.JSON(gox.M{"code": code, "data": data, "msg": msg})
}

// SendData sends a data response
func (r *APIResponder) SendData(data interface{}) {
	r.SendResponse(0, "", data)
}

// SendError sends an error response
func (r *APIResponder) SendError(err error) {
	var e *gox.Error
	if err == nil {
		e = gox.ErrSuccess
	} else {
		e = gox.ParseError(err)
		if e == nil {
			e = gox.NewError(gox.EcodeServer, err.Error())
		}
	}
	r.SendResponse(e.Code(), e.Msg(), nil)
	if e.Code() != 0 {
		log.Println("[WINE] SendError:", e, r.req)
	}
}

// SendMessage sends a message response
func (r *APIResponder) SendMessage(code int, msg string) {
	r.SendResponse(code, msg, nil)
	if code != 0 {
		log.Println("[WINE] Error:", code, msg, r.req)
	}
}

// SetLoginID sets current user id
func (r *APIResponder) SetLoginID(loginID gox.ID) {
	r.loginID = loginID
	log.Println("[WINE] Set user id[", loginID, "]", r.req.URL)
}

// LoginID return current user id
func (r *APIResponder) LoginID() gox.ID {
	return r.loginID
}
