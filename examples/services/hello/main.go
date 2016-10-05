package main

import (
	"fmt"
	"github.com/justintan/wine"
)

func main() {
	s := wine.Default()
	s.Get("/", func(c wine.Context) {
		c.Text("root")
	})

	s.Get("hi", func(c wine.Context) {
		c.Text("hi")
	})

	s.Group("hi").Get("/", func(c wine.Context) {
		c.Text("hi")
	})

	s.Get("hello", func(c wine.Context) {
		c.Text("Hello, wine!")
	})

	s.Get("docs/create", func(c wine.Context) {
		c.Text("Create doc")
	})

	s.Get("docs/:s/a", func(c wine.Context) {
		c.Text("Create doc: " + c.Params().GetStr("s"))
	})

	s.Get("docs/:doc_id", func(c wine.Context) {
		c.Text("doc id is " + c.Params().GetStr("doc_id"))
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
