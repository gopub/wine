package main

import (
	"context"
	"github.com/gopub/log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/wine/v2"
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
	s.Get("hello", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		responder.JSON(http.StatusOK, types.M{"code": 0, "msg": "hello"})
		return true
	})

	s.Get("login", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		username := request.Parameters().String("username")
		password := request.Parameters().String("password")
		if len(username) == 0 || len(password) == 0 {
			responder.JSON(http.StatusOK, types.M{"code": 1, "msg": "login error"})
			return true
		}
		u := &User{}
		u.ID = 1
		u.Name = "guest"
		u.CreatedAt = time.Now().Unix()
		responder.JSON(http.StatusOK, types.M{"code": 0, "msg": "success", "data": u})
		return true
	})

	s.Post("topic", func(ctx context.Context, request wine.Request, responder wine.Responder) bool {
		title := request.Parameters().String("title")
		if len(title) == 0 {
			responder.JSON(http.StatusOK, types.M{"code": 1, "msg": "no title"})
			return true
		}
		t := &Topic{}
		t.ID = 2
		t.User = &User{ID: 1, Name: "guest", CreatedAt: time.Now().Unix()}
		t.CreatedAt = time.Now().Unix()
		responder.JSON(http.StatusOK, types.M{"code": 2, "msg": "success", "data": t})
		return true
	})

	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	s.Run(":8000")
}
