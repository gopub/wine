package ws_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/gopub/wine"

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
	}).SetModel("")
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := ws.NewClient("ws://" + addr)
	var result string
	err := c.Call(context.Background(), "echo", "hello", &result)
	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

func TestHandshake(t *testing.T) {
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := ws.NewServer()
	s.Bind("echo", func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}).SetModel("")
	s.Handshake = func(rw ws.PacketReadWriter) error {
		p, err := rw.Read()
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		t.Logf("%s", wine.JSONString(p))
		err = rw.Write(p)
		assert.NoError(t, err)
		return err
	}
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := ws.NewClient("ws://" + addr)
	c.Handshake = func(rw ws.PacketReadWriter) error {
		data, err := json.Marshal(types.M{"greeting": "hello from tom"})
		require.NoError(t, err)
		p := new(ws.Packet)
		p.V = &ws.Packet_Push{
			Push: data,
		}
		err = rw.Write(p)
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		p, err = rw.Read()
		assert.NoError(t, err)
		t.Logf("%v", wine.JSONString(p))
		return err
	}
	var result string
	err := c.Call(context.Background(), "echo", "hello", &result)
	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

type AuthUserID int64

func (a AuthUserID) GetAuthUserID() int64 {
	return int64(a)
}

func TestServer_Push(t *testing.T) {
	var uid = AuthUserID(types.NewID().Int())
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := ws.NewServer()
	s.Bind("auth", func(ctx context.Context, req interface{}) (interface{}, error) {
		return uid, nil
	})
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := ws.NewClient("ws://" + addr)
	ctx := context.Background()
	err := c.Call(ctx, "auth", nil, nil)
	require.NoError(t, err)
	data := types.NewID().Pretty()
	err = s.Push(ctx, int64(uid), data)
	require.NoError(t, err)
	time.Sleep(time.Second) // Ensure client receive the data
	select {
	case pushData := <-c.PushDataC():
		var res string
		err = json.Unmarshal(pushData, &res)
		require.NoError(t, err)
		require.Equal(t, data, res)
	default:
		assert.Fail(t, "cannot recv push data")
	}
}
