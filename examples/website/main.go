package main

import "github.com/justintan/wine"

func main() {
	s := wine.Default()
	s.StaticDir("/", "/Users/qiyu/gocode/src/github.com/justintan/wine/examples/website/html")
	s.Run(":8000")
}
