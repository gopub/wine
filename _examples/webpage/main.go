package main

/*
 * Enter directory examples/websites/hello
 * go run ./main.go
 */

import (
	"github.com/gopub/wine"
)

func main() {
	s := wine.NewServer(nil)
	s.StaticDir("/", "./html")
	s.Run(":8000")
}
