package ws_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/gopub/types"
	"github.com/gopub/wine/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := ws.NewServer()
	s.Bind("echo", func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	})
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := ws.NewClient("ws://" + addr)
	result, err := c.Send(context.Background(), "echo", "hello")
	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

func TestHandshake(t *testing.T) {
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := ws.NewServer()
	s.Bind("echo", func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	})
	s.HandshakeHandler = func(rw ws.ReadWriter) error {
		var m types.M
		err := rw.ReadJSON(&m)
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		t.Logf("%v", m)
		err = rw.WriteJSON(types.M{"time": time.Now()})
		assert.NoError(t, err)
		return err
	}
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := ws.NewClient("ws://" + addr)
	c.HandshakeHandler = func(rw ws.ReadWriter) error {
		err := rw.WriteJSON(types.M{"greeting": "hello from tom"})
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		var m types.M
		err = rw.ReadJSON(&m)
		assert.NoError(t, err)
		t.Logf("%v", m)
		return err
	}
	result, err := c.Send(context.Background(), "echo", "hello")
	require.NoError(t, err)
	require.Equal(t, "hello", result)
}
