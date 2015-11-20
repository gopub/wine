package main

import (
	"fmt"
	"github.com/justintan/wine"
	"time"
)

func main() {

	s := wine.NewServer()

	//Intercept all requests with wine.Logger
	s.Use(wine.Logger)

	//Output html file
	s.StaticFile("/", "/var/www/index.html")

	//Map path dir to local dir
	s.StaticDir("/html/*", "/var/www/html")

	s.Get("users/:id/name", func(c wine.Context) {
		id := c.Params().GetStr("id")
		resp := map[string]interface{}{"name": "This is " + id + "'s name"}
		c.JSON(resp)
	})

	//Any means methods: GET POST PUT
	s.Any("server-time", func(c wine.Context) {
		resp := map[string]interface{}{"time": time.Now().Unix()}
		c.JSON(resp)
	})

	s.Post("update-name", auth, func(c wine.Context) {
		name := c.Params().GetStr("name")
		if len(name) == 0 {
			c.JSON(map[string]interface{}{"msg": "missing name"})
			return
		}
		c.JSON(map[string]interface{}{"msg": "new name is " + name})
	})

	s.Run(":8080")
}

func auth(c wine.Context) {
	sid := c.Params().GetStr("sid")
	fmt.Println(sid)
	//auth sid
	//...
	//simulate authorization
	if len(sid) > 0 {
		//authorized, call the next handler
		c.Next()
	} else {
		resp := map[string]interface{}{"msg": "authorization failed"}
		c.JSON(resp)
	}
}
