package main

import (
	"context"
	"github.com/gopub/log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/wine/v3"
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
	s := wine.DefaultServer()
	s.Get("hello", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		return wine.JSON(http.StatusOK, types.M{"code": 0, "msg": "hello"})
	})

	s.Get("login", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		username := req.Parameters.String("username")
		password := req.Parameters.String("password")
		if len(username) == 0 || len(password) == 0 {
			return wine.JSON(http.StatusOK, types.M{"code": 1, "msg": "login error"})
		}
		u := &User{}
		u.ID = 1
		u.Name = "guest"
		u.CreatedAt = time.Now().Unix()
		return wine.JSON(http.StatusOK, types.M{"code": 0, "msg": "success", "data": u})
	})

	s.Post("topic", func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		title := req.Parameters.String("title")
		if len(title) == 0 {
			return wine.JSON(http.StatusOK, types.M{"code": 1, "msg": "no title"})
		}
		t := &Topic{}
		t.ID = 2
		t.User = &User{ID: 1, Name: "guest", CreatedAt: time.Now().Unix()}
		t.CreatedAt = time.Now().Unix()
		return wine.JSON(http.StatusOK, types.M{"code": 2, "msg": "success", "data": t})
	})

	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	s.Run(":8000")
}
