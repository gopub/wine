package wine

import (
	"context"
	"github.com/gopub/log"
	"html/template"
	"net/http"
	"strings"

	"github.com/gopub/utils"
	"github.com/gopub/wine/v2/render"
)

// responderImpl is a default implementation of Context interface
type responderImpl struct {
	handlers  *handlerChain
	req       *http.Request
	writer    http.ResponseWriter
	responded bool
	templates []*template.Template
}

// Reset resets responder to be a new one
//func (dr *responderImpl) Reset(req *http.Request, rw http.ResponseWriter, tmpls []*template.Template) {
//	dr.responded = false
//	dr.req = req
//	dr.writer = rw
//	dr.templates = tmpls
//}

// Header returns response header
func (dr *responderImpl) Header() http.Header {
	return dr.writer.Header()
}

// Responded returns a flag to determine whether if the response has been written
func (dr *responderImpl) Responded() bool {
	return dr.responded
}

func (dr *responderImpl) markResponded() {
	if dr.responded {
		log.Panic("already responded")
	}
	dr.responded = true
}

// Send sends bytes
func (dr *responderImpl) Send(data []byte, contentType string) {
	dr.markResponded()
	if len(contentType) == 0 {
		contentType = http.DetectContentType(data)
	}
	if strings.Index(contentType, "charset") < 0 {
		contentType += "; charset=utf-8"
	}
	dr.Header()["Content-Type"] = []string{contentType}
	err := utils.WriteAll(dr.writer, data)
	if err != nil {
		log.Error("Send error:", err)
	}
}

// JSON sends json response
func (dr *responderImpl) JSON(status int, jsonObj interface{}) {
	dr.markResponded()
	render.JSON(dr.writer, status, jsonObj)
}

// Status sends a response just with a status code
func (dr *responderImpl) Status(status int) {
	dr.markResponded()
	render.Status(dr.writer, status)
}

// Redirect sends a redirect response
func (dr *responderImpl) Redirect(location string, permanent bool) {
	dr.writer.Header().Set("Location", location)
	if permanent {
		dr.Status(http.StatusMovedPermanently)
	} else {
		dr.Status(http.StatusFound)
	}
}

// File sends a file response
func (dr *responderImpl) File(filePath string) {
	dr.markResponded()
	http.ServeFile(dr.writer, dr.req, filePath)
}

// HTML sends a HTML response
func (dr *responderImpl) HTML(status int, htmlText string) {
	dr.markResponded()
	render.HTML(dr.writer, status, htmlText)
}

// Text sends a text response
func (dr *responderImpl) Text(status int, text string) {
	dr.markResponded()
	render.Text(dr.writer, status, text)
}

// TemplateHTML sends a HTML response. HTML page is rendered according to templateName and params
func (dr *responderImpl) TemplateHTML(templateName string, params interface{}) {
	for _, tmpl := range dr.templates {
		err := render.TemplateHTML(dr.writer, tmpl, templateName, params)
		if err == nil {
			dr.markResponded()
			break
		}
	}
}

// Handle handles request with h
func (dr *responderImpl) Handle(h http.Handler) {
	dr.markResponded()
	h.ServeHTTP(dr.writer, dr.req)
}

func (dr *responderImpl) Next(ctx context.Context, request Request, responder Responder) bool {
	h := dr.handlers.Next()
	if h == nil {
		log.Error("next handler is nil")
		return false
	}

	return h.HandleRequest(ctx, request, responder)
}
