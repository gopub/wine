package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gopub/log"

	"github.com/gopub/wine"
)

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type Topic struct {
	ID        int    `json:"id"`
	User      *User  `json:"user"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
}

func main() {
	s := wine.NewServer()
	s.Get("hello", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.JSON(http.StatusOK, map[string]interface{}{"code": 0, "msg": "hello"})
	})

	s.Get("login", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		username := req.Params().String("username")
		password := req.Params().String("password")
		if len(username) == 0 || len(password) == 0 {
			return wine.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": "login error"})
		}
		u := &User{}
		u.ID = 1
		u.Name = "guest"
		u.CreatedAt = time.Now().Unix()
		return wine.JSON(http.StatusOK, map[string]interface{}{"code": 0, "msg": "success", "data": u})
	})

	s.Post("topic", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		title := req.Params().String("title")
		if len(title) == 0 {
			return wine.JSON(http.StatusOK, map[string]interface{}{"code": 1, "msg": "no title"})
		}
		t := &Topic{}
		t.ID = 2
		t.User = &User{ID: 1, Name: "guest", CreatedAt: time.Now().Unix()}
		t.CreatedAt = time.Now().Unix()
		return wine.JSON(http.StatusOK, map[string]interface{}{"code": 2, "msg": "success", "data": t})
	})

	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	s.Run(":8000")
}
