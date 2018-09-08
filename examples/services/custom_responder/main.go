package main

import (
	"github.com/gopub/wine"
	"time"
)

func main() {
	s := wine.DefaultServer()
	s.RegisterResponder(&APIResponder{})
	s.Get("time", func(c *wine.Context) {
		r := c.Responder.(*APIResponder)
		r.SendResponse(0, "", time.Now().Unix())
	})
	s.Run(":8000")
}
