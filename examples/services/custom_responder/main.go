package main

import (
	"time"

	"github.com/gopub/wine"
)

func main() {
	s := wine.DefaultServer()
	s.RegisterResponder(&wine.APIResponder{})
	s.Get("time", func(c *wine.Context) {
		r := c.Responder.(*wine.APIResponder)
		r.SendResponse(0, "", time.Now().Unix())
	})
	s.Run(":8000")
}
