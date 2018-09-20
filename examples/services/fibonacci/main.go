package main

import (
	"context"
	"github.com/gopub/types"
	"github.com/gopub/wine/v2"
	"net/http"
)

func main() {
	s := wine.DefaultServer()
	s.Get("/fibonacci", func(ctx context.Context, request wine.Request, invoker wine.Invoker) wine.Responsible {
		n := request.Parameters().Int("n")
		result := fibonacci(n)
		return wine.JSON(http.StatusOK, types.M{"result": result})
	})
	s.Run(":8000")
}

func fibonacci(n int) int {
	if n < 3 {
		return 1
	}

	return fibonacci(n-1) + fibonacci(n-2)
}
