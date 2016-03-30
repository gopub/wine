package main

import (
	"github.com/justintan/wine"
	"log"
	"strconv"
	"time"
)

func Logger(c wine.Context) {
	st := time.Now()
	c.Next()
	cost := float32((time.Since(st) / time.Microsecond)) / 1000.0
	req := c.HTTPRequest()
	log.Printf("[WINE] %.3fms %s %s", cost, req.Method, req.RequestURI)
}

func main() {
	s := wine.NewServer()
	s.Use(Logger)
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
