package wine

import "github.com/justintan/gox"

type Context interface {
	Request() *Request
	Session() Session
	NewSession() Session
	SendResponse(resp *Response)
	SendCode(code gox.Code)
	SendData(data gox.M)
}
