package main

import (
	"github.com/justintan/wine"
	"log"
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
	s := wine.Default()
	s.Post("feedback", func(c wine.Context) {
		text := c.Params().GetStr("text")
		email := c.Params().GetStr("email")
		c.Text("Feedback:" + text + " from " + email)
	})
	s.Run(":8000")
}
