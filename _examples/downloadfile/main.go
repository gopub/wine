package main

import (
	"context"

	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer()
	s.Get("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		data := []byte("Hello, world!")
		return wine.BytesFile(data, "test.txt")
	})
	s.Run(":8000")
}
