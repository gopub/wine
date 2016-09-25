package main

import (
	"github.com/justintan/gox/types"
	"github.com/justintan/wine"
	"net/http"
	_ "net/http/pprof"
	"qiniupkg.com/x/log.v7"
	"time"
)

type User struct {
	ID        types.ID `json:"id"`
	Name      string   `json:"name"`
	CreatedAt int64    `json:"created_at"`
}

type Topic struct {
	ID        types.ID `json:"id"`
	User      *User    `json:"user"`
	Title     string   `json:"title"`
	CreatedAt int64    `json:"created_at"`
}

func main() {
	s := wine.Default()
	s.Get("login", func(c wine.Context) {
		username := c.Params().GetStr("username")
		password := c.Params().GetStr("password")
		if len(username) == 0 || len(password) == 0 {
			c.JSON(types.M{"code": 1, "msg": "login error"})
			return
		}
		u := &User{}
		u.ID = 1
		u.Name = "guest"
		u.CreatedAt = time.Now().Unix()
		c.JSON(types.M{"code": 0, "msg": "success", "data": u})
	})

	s.Post("topic", func(c wine.Context) {
		title := c.Params().GetStr("title")
		if len(title) == 0 {
			c.JSON(types.M{"code": 1, "msg": "no title"})
			return
		}
		t := &Topic{}
		t.ID = 2
		t.User = &User{ID: 1, Name: "guest", CreatedAt: time.Now().Unix()}
		t.CreatedAt = time.Now().Unix()
		c.JSON(types.M{"code": 2, "msg": "success", "data": t})
	})

	go func() {
		log.Debug(http.ListenAndServe("localhost:6060", nil))
	}()

	s.Run(":8000")
}
