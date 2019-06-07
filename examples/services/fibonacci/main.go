package main

import (
	"context"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer(wine.DefaultConfig())
	s.Get("/fibonacci", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		n := req.Parameters.Int("n")
		result := fibonacci(n)
		return wine.JSON(http.StatusOK, gox.M{"result": result})
	})
	s.Run(":8000")
}

func fibonacci(n int) int {
	if n < 3 {
		return 1
	}

	return fibonacci(n-1) + fibonacci(n-2)
}
