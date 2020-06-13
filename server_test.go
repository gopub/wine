package wine_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"testing"

	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
	"github.com/stretchr/testify/require"
)

func TestServerStatus(t *testing.T) {
	server := wine.NewServer()
	r := server.Router
	r.Get("/ok", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.OK
	})
	r.Get("/forbidden", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Status(http.StatusForbidden)
	})
	addr := fmt.Sprintf("localhost:%d", rand.Int()%1000+8000)
	host := "http://" + addr
	go server.Run(addr)

	t.Run("OK", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(host + "/ok")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Forbidden", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(host + "/forbidden")
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(host + "/notfoundHandler")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestServerMethod(t *testing.T) {
	server := wine.NewServer()
	r := server.Router
	r.Get("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, "GET")
	})
	r.Post("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, "POST")
	})
	r.Put("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, "PUT")
	})
	addr := fmt.Sprintf("localhost:%d", rand.Int()%1000+8000)
	host := "http://" + addr
	go server.Run(addr)
	runtime.Gosched()
	t.Run("GET", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(host)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "GET", string(body))
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := http.DefaultClient.Post(host, mime.Plain, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "POST", string(body))
	})

	t.Run("PUT", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, host, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "PUT", string(body))
	})

	t.Run("NotFound", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, host, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
