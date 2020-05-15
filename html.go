package wine

import (
	"context"
	"net/http"

	"github.com/gopub/wine/mime"
)

// HTML creates a HTML response
func HTML(status int, html string) *Response {
	header := make(http.Header)
	header.Set(mime.ContentType, mime.HtmlUTF8)
	return &Response{
		status: status,
		header: header,
		value:  html,
	}
}

// TemplateHTML sends a HTML response. HTML page is rendered according to name and params
func TemplateHTML(name string, params interface{}) Responder {
	return ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		for _, tmpl := range GetTemplates(ctx) {
			var err error
			if name == "" {
				err = tmpl.Execute(w, params)
			} else {
				err = tmpl.ExecuteTemplate(w, name, params)
			}

			if err == nil {
				break
			}
		}
	})
}
