package main

import (
	"github.com/justintan/wine"
	"strconv"
	"time"
)

func main() {
	s := wine.Default()
	s.Get("/hello", func(c wine.Context) {
		c.Text("Hello, Wine!")
	})

	s.Get("/time", func(c wine.Context) {
		c.JSON(map[string]interface{}{"time": time.Now().Unix()})
	})

	s.Get("/items/:id", func(c wine.Context) {
		id := c.Params().GetStr("id")
		c.Text("item id: " + id)
	})

	s.Get("/items/:page,:size", func(c wine.Context) {
		page := c.Params().GetInt("page")
		size := c.Params().GetInt("size")
		c.Text("page:" + strconv.Itoa(page) + " size:" + strconv.Itoa(size))
	})

	s.Run(":8000")
}
