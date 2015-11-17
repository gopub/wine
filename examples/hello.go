package main

import (
	"fmt"
	"github.com/justintan/wine"
	"time"
)

func main() {

	s := wine.Server()

	//You can implement middle wares and add them to the routing
	s.Use(wine.Logger)
	s.Use()

	//support file and directory
	s.StaticFile("/", "/var/www/index.html")
	s.StaticDir("/html/*", "/var/www/html")

	s.Get("server-time", func(c wine.Context) {
		resp := map[string]interface{}{"time": time.Now().Unix()}
		c.SendJSON(resp)
	})

	s.Get("users/:id/name", func(c wine.Context) {
		id := c.RequestParams().GetStr("id")
		resp := map[string]interface{}{"name": "This is " + id + "'s name"}
		c.SendJSON(resp)
	})

	//Equals to s.Get("login", login) and s.Post("login", login)
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
		c.SendJSON(resp)
	}
}

func login(c wine.Context) {
	account := c.RequestParams().GetStr("account")
	password := c.RequestParams().GetStr("password")
	fmt.Println(account, password)
	resp := map[string]interface{}{"status": "success"}
	c.SendJSON(resp)
}

func getProfile(c wine.Context) {
	id := c.RequestParams().GetStr("id")
	resp := map[string]interface{}{"profile": "This is " + id + "'s profile"}
	c.SendJSON(resp)
}

func updateName(c wine.Context) {
	name := c.RequestParams().GetStr("name")
	resp := map[string]interface{}{"name": name}
	c.SendJSON(resp)
}
