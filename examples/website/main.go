package main

import (
	"github.com/justintan/wine"
)

func main() {
	s := wine.Default()
	s.StaticDir("/", "./html")
	s.Run(":8000")
}
