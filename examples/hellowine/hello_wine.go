package main

import "github.com/justintan/wine"

func main() {
	s := wine.Server()
	s.GP("/", func(c wine.Context) {
		c.SendHTML("Hello, This is WINE!")
	})
	s.Run(":8001")
}
