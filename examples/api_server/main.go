package main

import (
	"github.com/justintan/gox"
	"github.com/justintan/wine"
	"time"
)

func main() {
	s := wine.Default()

	s.Get("whattime", func(c wine.Context) {
		c.JSON(gox.M{"time": time.Now()})
	})

	s.Get("list/:page,:size", func(c wine.Context) {
		c.HTML(c.Params().GetStr("page") + ", " + c.Params().GetStr("size"))
	})

	s.Get("users/:user_id/name", func(c wine.Context) {
		c.HTML(c.Params().GetStr("user_id") + "'s name is Wine")
	})

	s.Post("users/:user_id/name/:name", func(c wine.Context) {
		c.HTML(c.Params().GetStr("user_id") + "'s new name is " + c.Params().GetStr("name"))
	})

	s.Any("login", func(c wine.Context) {
		username := c.Params().GetStr("username")
		password := c.Params().GetStr("password")
		gox.LDebug(username, password)
		c.JSON(gox.M{"status": 0, "token": gox.NewUUID(), "msg": "success"})
	})

	s.Run(":8000")
}
