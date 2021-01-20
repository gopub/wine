package wine_test

import (
	"context"
	"fmt"
	"github.com/gopub/errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/gopub/wine"
	"github.com/gopub/wine/httpvalue"
	"github.com/stretchr/testify/require"
)

func TestServerStatus(t *testing.T) {
	server := wine.NewTestServer()
	r := server.Router
	r.Get("/ok", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.OK
	})
	r.Get("/forbidden", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Status(http.StatusForbidden)
	})
	url := server.Run()
	t.Run("OK", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(url + "/ok")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Forbidden", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(url + "/forbidden")
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("NotFound", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(url + "/notfoundHandler")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestServerMethod(t *testing.T) {
	server := wine.NewTestServer()
	r := server.Router
	getStr := strings.Repeat(uuid.New().String(), 1024)
	r.Get("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, getStr)
	})
	r.Post("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, "POST")
	})
	r.Put("/", func(ctx context.Context, req *wine.Request) wine.Responder {
		return wine.Text(http.StatusOK, "PUT")
	})
	url := server.Run()
	runtime.Gosched()
	t.Run("GET", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(url)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		require.Equal(t, getStr, string(body))
	})

	t.Run("POST", func(t *testing.T) {
		resp, err := http.DefaultClient.Post(url, httpvalue.Plain, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, resp.Body.Close())
		require.NoError(t, err)
		require.Equal(t, "POST", string(body))
	})

	t.Run("PUT", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPut, url, nil)
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
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestServer_Header(t *testing.T) {
	server := wine.NewTestServer()
	url := server.Run()
	t.Run("RootHeader", func(t *testing.T) {
		v := uuid.New().String()
		key := "RootHeader"
		server.Header().Add(key, v)
		resp, err := http.Get(url)
		require.NoError(t, err)
		require.Equal(t, v, resp.Header.Get(key))
	})
	t.Run("PathHeader", func(t *testing.T) {
		v := uuid.New().String()
		key := "PathHeader"
		server.Get("/hello", func(ctx context.Context, req *wine.Request) wine.Responder {
			return wine.OK
		}).Header().Add(key, v)
		resp, err := http.Get(url + "/hello")
		require.NoError(t, err)
		require.Equal(t, v, resp.Header.Get(key))
	})
}

func TestServer_Bind(t *testing.T) {
	server := wine.NewTestServer()
	url := server.Run()
	t.Run("PathParamInt64", func(t *testing.T) {
		id := rand.Int63()
		server.Get("int/{id}", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(int64) == id {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel(int64(0))
		resp, err := http.Get(url + "/int/" + fmt.Sprint(id))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("PathParamInt64Error", func(t *testing.T) {
		id := rand.Int63()
		server.Get("interror/{id}", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(int64) == id {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel(int64(0))
		resp, err := http.Get(url + "/interror/" + uuid.New().String())
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("PathParamString", func(t *testing.T) {
		name := uuid.New().String()
		server.Get("/string/{name}", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(string) == name {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel("")
		resp, err := http.Get(url + "/string/" + name)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("MultiPathParamInt64Error", func(t *testing.T) {
		id := rand.Int63()
		server.Get("/multi/{name}/{id}", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(int64) == id {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel(int64(0))
		resp, err := http.Get(url + "/multi/haha/" + fmt.Sprint(id))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("QueryParamInt64", func(t *testing.T) {
		id := rand.Int63()
		server.Get("query/int", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(int64) == id {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel(int64(0))
		resp, err := http.Get(url + "/query/int?id=" + fmt.Sprint(id))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("MultiQueryParamInt64Error", func(t *testing.T) {
		id := rand.Int63()
		server.Get("multiquery/int", func(ctx context.Context, req *wine.Request) wine.Responder {
			if req.Model.(int64) == id {
				return wine.OK
			}
			return errors.BadRequest("")
		}).SetModel(int64(0))
		resp, err := http.Get(url + "/multiquery/int?name=haha&id=" + fmt.Sprint(id))
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
