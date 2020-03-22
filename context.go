package wine

import (
	"context"
	"html/template"
	"net/http"
)

type contextKey int

// Context keys
const (
	ckBasicAuthUser contextKey = iota + 1
	ckHTTPResponseWriter
	ckTemplates
	ckSessionID
)

func GetSessionID(ctx context.Context) string {
	v := ctx.Value(ckSessionID)
	sid, _ := v.(string)
	return sid
}

func withSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, ckSessionID, sid)
}

// GetResponseWriter returns response writer from the context
func GetResponseWriter(ctx context.Context) http.ResponseWriter {
	rw, _ := ctx.Value(ckHTTPResponseWriter).(http.ResponseWriter)
	return rw
}

func withResponseWriter(ctx context.Context, rw http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ckHTTPResponseWriter, rw)
}

// GetTemplates returns templates in context
func GetTemplates(ctx context.Context) []*template.Template {
	v, _ := ctx.Value(ckTemplates).([]*template.Template)
	return v
}

func withTemplate(ctx context.Context, templates []*template.Template) context.Context {
	return context.WithValue(ctx, ckTemplates, templates)
}
