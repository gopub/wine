package main

import (
	"fmt"
	"github.com/justintan/wine"
)

func main() {
	s := wine.Default()
	s.Get("hello", func(c wine.Context) {
		c.Text("Hello, wine!")
	})

	s.Get("sum/:a,:b", func(c wine.Context) {
		r := c.Params().GetInt("a") + c.Params().GetInt("b")
		c.Text(fmt.Sprint(r))
	})

	s.Get("sum/:a,:b/hehe", func(c wine.Context) {
		r := c.Params().GetInt("a") * c.Params().GetInt("b")
		c.Text(fmt.Sprint(r))
	})

	s.Get("sum/:a,:b,:c", func(c wine.Context) {
		r := c.Params().GetInt("a") + c.Params().GetInt("b") + c.Params().GetInt("c")
		c.Text(fmt.Sprint(r))
	})

	s.StaticDir("hello/*", "../../websites/hello/html")

	s.Run(":8000")
}
