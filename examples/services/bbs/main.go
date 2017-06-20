package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/natande/gox"
	"github.com/natande/wine"
)

type User struct {
	ID        gox.ID `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type Topic struct {
	ID        gox.ID `json:"id"`
	User      *User  `json:"user"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
}

func main() {
	s := wine.DefaultServer()
	s.Get("hello", func(c wine.Context) {
		c.JSON(gox.M{"code": 0, "msg": "hello"})
	})

	s.Get("login", func(c wine.Context) {
		username := c.Params().String("username")
		password := c.Params().String("password")
		if len(username) == 0 || len(password) == 0 {
			c.JSON(gox.M{"code": 1, "msg": "login error"})
			return
		}
		u := &User{}
		u.ID = 1
		u.Name = "guest"
		u.CreatedAt = time.Now().Unix()
		c.JSON(gox.M{"code": 0, "msg": "success", "data": u})
	})

	s.Post("topic", func(c wine.Context) {
		title := c.Params().String("title")
		if len(title) == 0 {
			c.JSON(gox.M{"code": 1, "msg": "no title"})
			return
		}
		t := &Topic{}
		t.ID = 2
		t.User = &User{ID: 1, Name: "guest", CreatedAt: time.Now().Unix()}
		t.CreatedAt = time.Now().Unix()
		c.JSON(gox.M{"code": 2, "msg": "success", "data": t})
	})

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	s.Run(":8000")
}
