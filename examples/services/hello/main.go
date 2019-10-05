package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gopub/wine/api"

	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer()
	s.Get("echo", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return api.Data(req.Params())
	})

	s.Get("/", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "root")
	})

	s.Get("hi", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "hi")
	})

	s.Get("hello", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Hello, wine!")
	})

	s.Get("docs/create", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Create doc")
	})

	s.Get("docs/{s}/a", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Create doc: "+req.Params().String("s"))
	})

	s.Get("docs/{doc_id}", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "doc id is "+req.Params().String("doc_id"))
	})

	s.Get("sum/{a}/{b}", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		r := req.Params().Int("a") + req.Params().Int("b")
		return wine.Text(http.StatusOK, fmt.Sprint(r))
	})

	s.Get("sum/{a}/{b}/hehe", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		r := req.Params().Int("a") * req.Params().Int("b")
		return wine.Text(http.StatusOK, fmt.Sprint(r))
	})

	s.StaticDir("hello/*", "../../websites/hello/html")

	s.Run(":8000")
}
