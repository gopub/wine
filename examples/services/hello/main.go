package main

import "github.com/justintan/wine"

func main() {
	s := wine.Default()
	s.Get("hello", func(c wine.Context) {
		c.Text("Hello, wine!")
	})
	s.Run(":8000")
}
