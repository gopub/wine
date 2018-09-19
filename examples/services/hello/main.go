package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gopub/wine/v2"
)

func main() {
	s := wine.DefaultServer()
	s.Get("/", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "root")
		return true
	})

	s.Get("hi", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "hi")
		return true
	})

	s.Get("hello", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "Hello, wine!")
		return true
	})

	s.Get("docs/create", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "Create doc")
		return true
	})

	s.Get("docs/:s/a", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "Create doc: "+request.Parameters().String("s"))
		return true
	})

	s.Get("docs/:doc_id", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.Text(http.StatusOK, "doc id is "+request.Parameters().String("doc_id"))
		return true
	})

	s.Get("sum/:a,:b", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		r := request.Parameters().Int("a") + request.Parameters().Int("b")
		responder.Text(http.StatusOK, fmt.Sprint(r))
		return true
	})

	s.Get("sum/:a,:b/hehe", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		r := request.Parameters().Int("a") * request.Parameters().Int("b")
		responder.Text(http.StatusOK, fmt.Sprint(r))
		return true
	})

	s.Get("sum/:a,:b,:c", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		r := request.Parameters().Int("a") + request.Parameters().Int("b") + request.Parameters().Int("c")
		responder.Text(http.StatusOK, fmt.Sprint(r))
		return true
	})

	s.StaticDir("hello/*", "../../websites/hello/html")

	s.Run(":8000")
}
