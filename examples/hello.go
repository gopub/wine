package main

import (
	"fmt"
	"github.com/justintan/wine"
	"time"
)

func main() {

	s := wine.NewServer()

	//You can implement middlewares and add them to the server
	s.Use(wine.Logger)
	s.Use()

	//Support file and directory
	s.StaticFile("/", "/var/www/index.html")
	s.StaticDir("/html/*", "/var/www/html")

	s.Get("server-time", func(c wine.Context) {
		resp := map[string]interface{}{"time": time.Now().Unix()}
		c.JSON(resp)
	})

	s.Get("users/:id/name", func(c wine.Context) {
		id := c.Params().GetStr("id")
		resp := map[string]interface{}{"name": "This is " + id + "'s name"}
		c.JSON(resp)
	})

	//Means accept methods: GET POST PUT DELETE
	s.Any("login", login)

	s.Run(":8080")
}

func auth(c wine.Context) {
	sid := c.Get("session_id")
	fmt.Println(sid)
	//auth sid
	//...
	authorized := false

	if authorized {
		//call the next handler
		c.Next()
	} else {
		//abort the handling process, send an error response
		resp := map[string]interface{}{"msg": "authorization failed"}
		c.JSON(resp)
	}
}

func login(c wine.Context) {
	account := c.Params().GetStr("account")
	password := c.Params().GetStr("password")
	fmt.Println(account, password)
	resp := map[string]interface{}{"status": "success"}
	c.JSON(resp)
}

func getProfile(c wine.Context) {
	id := c.Params().GetStr("id")
	resp := map[string]interface{}{"profile": "This is " + id + "'s profile"}
	c.JSON(resp)
}

func updateName(c wine.Context) {
	name := c.Params().GetStr("name")
	resp := map[string]interface{}{"name": name}
	c.JSON(resp)
}
