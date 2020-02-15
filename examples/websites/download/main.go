package main

import (
	"context"
	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer()
	s.Get("download", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		data := []byte("Hello, world!")
		return wine.File(data, "test.txt")
	})
	s.Run(":8000")
}
