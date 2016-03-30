package wine

import (
	"github.com/justintan/gox/types"
	"github.com/justintan/wine/render"
	"html/template"
	"net/http"
	"strings"
)

type DefaultContext struct {
	keyValues types.M
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
	handlers  *HandlerChain

	req        *http.Request
	reqHeader  http.Header
	reqParams  types.M
	respHeader http.Header
}

func (dc *DefaultContext) Rebuild(rw http.ResponseWriter, req *http.Request, templates []*template.Template, handlers []Handler) {
	if dc.keyValues != nil {
		for k := range dc.keyValues {
			delete(dc.keyValues, k)
		}
	} else {
		dc.keyValues = types.M{}
	}

	if dc.reqHeader != nil {
		for k := range dc.reqHeader {
			delete(dc.reqHeader, k)
		}
	} else {
		dc.reqHeader = make(http.Header)
	}

	if dc.respHeader != nil {
		for k := range dc.respHeader {
			delete(dc.respHeader, k)
		}
	} else {
		dc.respHeader = make(http.Header)
	}

	dc.responded = false
	dc.writer = rw
	dc.req = req
	dc.reqParams = parseHTTPReq(req)
	dc.handlers = NewHandlerChain(handlers)
	dc.templates = templates

	for k, v := range req.Header {
		k = strings.ToLower(k)
		if strings.Index(k, "x-") == 0 {
			k = k[2:]
		}
		dc.reqHeader[k] = v
	}
}

func (dc *DefaultContext) Set(key string, value interface{}) {
	dc.keyValues[key] = value
}

func (dc *DefaultContext) Get(key string) interface{} {
	return dc.keyValues[key]
}

func (dc *DefaultContext) Next() {
	if h := dc.handlers.Next(); h != nil {
		h.HandleRequest(dc)
	}
}

func (dc *DefaultContext) HTTPRequest() *http.Request {
	return dc.req
}

func (dc *DefaultContext) Params() types.M {
	return dc.reqParams
}

func (dc *DefaultContext) Header() http.Header {
	return dc.reqHeader
}

func (dc *DefaultContext) ResponseHeader() http.Header {
	return dc.respHeader
}

func (dc *DefaultContext) Responded() bool {
	return dc.responded
}

func (dc *DefaultContext) setResponded() {
	if dc.responded {
		panic("[WINE] already responded")
	}
	dc.responded = true
}

func (dc *DefaultContext) JSON(jsonObj interface{}) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.JSON(dc.writer, jsonObj)
}

func (dc *DefaultContext) Status(status int) {
	dc.setResponded()
	for k, v := range dc.respHeader {
		dc.writer.Header()[k] = v
	}
	render.Status(dc.writer, status)
}

func (dc *DefaultContext) File(filePath string) {
	dc.setResponded()
	http.ServeFile(dc.writer, dc.req, filePath)
}

func (dc *DefaultContext) HTML(htmlText string) {
	dc.setResponded()
	render.HTML(dc.writer, htmlText)
}

func (dc *DefaultContext) TemplateHTML(templateName string, params types.M) {
	for _, tpl := range dc.templates {
		err := render.TemplateHTML(dc.writer, tpl, templateName, params)
		if err == nil {
			dc.setResponded()
			break
		}
	}
}

func (dc *DefaultContext) ServeHTTP(h http.Handler) {
	dc.setResponded()
	h.ServeHTTP(dc.writer, dc.req)
}
