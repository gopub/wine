package main

/*
 * Enter directory examples/websites/hello
 * go run ./main.go
 */

import "github.com/gopub/wine/v3"

func main() {
	s := wine.DefaultServer()
	s.StaticDir("/", "./html")
	s.Run(":8000")
}
