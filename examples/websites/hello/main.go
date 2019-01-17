package main

/*
 * Enter directory examples/websites/hello
 * go run ./main.go
 */

import (
	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer(wine.DefaultConfig())
	s.StaticDir("/", "./html")
	s.Run(":8000")
}
