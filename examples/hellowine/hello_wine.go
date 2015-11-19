package main

import "github.com/justintan/wine"

func main() {
	s := wine.Server()
	s.Any("/", func(c wine.Context) {
		c.HTML("Hello, This is WINE!")
	})
	s.Run(":8001")
}
