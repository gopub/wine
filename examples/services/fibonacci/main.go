package main

import (
	"github.com/gopub/types"
	"github.com/gopub/wine"
	"net/http"
)

func main() {
	s := wine.DefaultServer()
	s.Get("/fibonacci", func(c *wine.Context) {
		n := c.Params().Int("n")
		result := fibonacci(n)
		c.JSON(http.StatusOK, types.M{"result": result})
	})
	s.Run(":8000")
}

func fibonacci(n int) int {
	if n < 3 {
		return 1
	}

	return fibonacci(n-1) + fibonacci(n-2)
}
