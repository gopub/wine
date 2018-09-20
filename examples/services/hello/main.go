package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gopub/wine/v3"
)

func main() {
	s := wine.DefaultServer()
	s.Get("/", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "root")
	})

	s.Get("hi", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "hi")
	})

	s.Get("hello", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Hello, wine!")
	})

	s.Get("docs/create", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Create doc")
	})

	s.Get("docs/:s/a", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "Create doc: "+request.Parameters().String("s"))
	})

	s.Get("docs/:doc_id", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		return wine.Text(http.StatusOK, "doc id is "+request.Parameters().String("doc_id"))
	})

	s.Get("sum/:a,:b", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		r := request.Parameters().Int("a") + request.Parameters().Int("b")
		return wine.Text(http.StatusOK, fmt.Sprint(r))
	})

	s.Get("sum/:a,:b/hehe", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		r := request.Parameters().Int("a") * request.Parameters().Int("b")
		return wine.Text(http.StatusOK, fmt.Sprint(r))
	})

	s.Get("sum/:a,:b,:c", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		r := request.Parameters().Int("a") + request.Parameters().Int("b") + request.Parameters().Int("c")
		return wine.Text(http.StatusOK, fmt.Sprint(r))
	})

	s.StaticDir("hello/*", "../../websites/hello/html")

	s.Run(":8000")
}
