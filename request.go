package wine

import (
	"fmt"
	"github.com/justintan/gox"
	"net/http"
)

const HeaderPrefix = "wine-"

type Request struct {
	API    string `json:"api"`
	Header gox.M  `json:"header"`
	Data   gox.M  `json:"data"`
	Method string `json:"method"`
}

func NewRequest(api string, header gox.M, data gox.M) *Request {
	if header == nil {
		header = gox.M{}
	}

	if data == nil {
		data = gox.M{}
	}

	return &Request{API: api, Header: header, Data: data, Method: "GET"}
}

func SendRequest(destUrl string, req *Request) *Response {
	header := http.Header{}
	for k, v := range req.Header {
		header.Add(k, fmt.Sprint(v))
	}

	result := gox.FetchJSON(destUrl, string(req.Method), header, req.Data)
	if len(result) == 0 {
		return NewResponse(gox.BadRequest, "", nil)
	}

	resp := NewResponse(gox.OK, "", nil)
	resp.Data, _ = result["data"].(map[string]interface{})
	if resp.Data == nil {
		resp.Data = gox.M{}
	}

	resp.Code = gox.Code(result.GetInt("code"))
	resp.Msg = result.GetStr("msg")
	gox.Log().Debug("[RequestData]", destUrl, req, "result:", resp)
	return resp
}

func (this *Request) String() string {
	return fmt.Sprint(this.API, this.Header, this.Method, this.Data)
}
