package main

import (
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"html/template"
	"net/http"

	"github.com/gopub/types"
)

var _ wine.Responder = (*APIResponder)(nil)

//APIResponder is designed for api request/response
type APIResponder struct {
	*wine.DefaultResponder
	loginID types.ID
}

// Reset resets the Responder
func (ar *APIResponder) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
	if ar.DefaultResponder == nil {
		ar.DefaultResponder = &wine.DefaultResponder{}
	}
	ar.DefaultResponder.Reset(req, rw, tmpls)
}

// SendResponse sends a response
func (ar *APIResponder) SendResponse(code int, msg string, data interface{}) {
	ar.JSON(http.StatusOK, types.M{"code": code, "data": data, "msg": msg})
}

// SendData sends a data response
func (ar *APIResponder) SendData(data interface{}) {
	ar.SendResponse(0, "", data)
}

// SendMessage sends a message response
func (ar *APIResponder) SendMessage(code int, msg string) {
	ar.SendResponse(code, msg, nil)
	if code != 0 {
		log.Error(code, msg)
	}
}

// SetLoginID sets current login user id
func (ar *APIResponder) SetLoginID(loginID types.ID) {
	ar.loginID = loginID
	log.Infof("loginID=%d", loginID)
}

// LoginID return current login user id
func (ar *APIResponder) LoginID() types.ID {
	return ar.loginID
}
