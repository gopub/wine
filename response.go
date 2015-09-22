package wine

import "github.com/justintan/gox"

type Response struct {
	Code gox.Code `json:"code"`
	Msg  string   `json:"msg,omitempty"`
	Data gox.M    `json:"data,omitempty"`
}

func (this *Response) ToMap() gox.M {
	return gox.M{"code": this.Code, "msg": this.Msg, "data": this.Data}
}

func NewResponse(code gox.Code, msg string, data gox.M) *Response {
	return &Response{Code: code, Msg: msg, Data: data}
}
