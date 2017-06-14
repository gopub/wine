package main

import (
	"github.com/justintan/gox/types"
	"github.com/justintan/wine"
)

func main() {
	s := wine.DefaultServer()
	s.Get("/fibonacci", func(c wine.Context) {
		n := c.Params().Int("n")
		result := fibonacci(n)
		c.JSON(types.M{"result": result})
	})
	s.Run(":8000")
}

func fibonacci(n int) int {
	if n < 3 {
		return 1
	}

	return fibonacci(n-1) + fibonacci(n-2)
}
