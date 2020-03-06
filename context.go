package wine

import (
	"context"
	"html/template"
	"net/http"

	"github.com/gopub/gox"
)

// Context keys
const (
	CKBasicAuthUser      gox.ContextKey = "wine_basic_auth_user"
	CKHTTPResponseWriter gox.ContextKey = "wine_http_resp_writer"
	CKTemplates          gox.ContextKey = "wine_templates"
	CKSessionID          gox.ContextKey = "wine_sid"
)

func GetSessionID(ctx context.Context) string {
	v := ctx.Value(CKSessionID)
	sid, _ := v.(string)
	return sid
}

func withSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, CKSessionID, sid)
}

// GetResponseWriter returns response writer from the context
func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(CKHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func withResponseWriter(ctx context.Context, rw http.ResponseWriter) context.Context {
	return context.WithValue(ctx, CKHTTPResponseWriter, rw)
}

// GetTemplates returns templates in context
func GetTemplates(ctx context.Context) []*template.Template {
	v, _ := ctx.Value(CKTemplates).([]*template.Template)
	return v
}

func withTemplate(ctx context.Context, templates []*template.Template) context.Context {
	return context.WithValue(ctx, CKTemplates, templates)
}
