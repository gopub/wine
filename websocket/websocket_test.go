package websocket_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/gopub/conv"
	"github.com/gopub/types"
	"github.com/gopub/wine/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Send(t *testing.T) {
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := websocket.NewServer()
	s.Bind("echo", func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}).SetModel("")
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := websocket.NewClient("ws://"+addr, nil)
	var result string
	err := c.Call(context.Background(), "echo", "hello", &result)
	require.NoError(t, err)
	require.Equal(t, "hello", result)
}

func TestHandshake(t *testing.T) {
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := websocket.NewServer()
	s.Bind("echo", func(ctx context.Context, req interface{}) (interface{}, error) {
		return req, nil
	}).SetModel("")
	s.Handshake = func(rw websocket.PacketReadWriter) error {
		p, err := rw.Read()
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		err = rw.Write(p)
		assert.NoError(t, err)
		return err
	}
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := websocket.NewClient("ws://"+addr, nil)
	c.Handshaker = func(rw websocket.PacketReadWriter) error {
		data, err := json.Marshal(types.M{"greeting": "hello from tom"})
		require.NoError(t, err)
		p := new(websocket.Packet)
		p.V = &websocket.Packet_Data{
			Data: &websocket.Data{
				V: &websocket.Data_Json{Json: data},
			},
		}
		err = rw.Write(p)
		assert.NoError(t, err)
		if err != nil {
			return err
		}
		p, err = rw.Read()
		assert.NoError(t, err)
		t.Logf("%v", conv.MustJSONString(p))
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

func (a AuthUserID) GetConnID() string {
	return fmt.Sprint(a)
}

func TestServer_Push(t *testing.T) {
	var uid = AuthUserID(types.NewID().Int())
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := websocket.NewServer()
	s.Bind("auth", func(ctx context.Context, req interface{}) (interface{}, error) {
		return uid, nil
	})
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := websocket.NewClient("ws://"+addr, nil)
	ctx := context.Background()
	err := c.Call(ctx, "auth", nil, nil)
	require.NoError(t, err)
	data := types.NewID().Pretty()
	err = s.Push(ctx, fmt.Sprint(uid), 10, data)
	require.NoError(t, err)
	time.Sleep(time.Second) // Ensure client receive the data
	select {
	case push := <-c.PushC():
		var res string
		err = push.Data.Unmarshal(&res)
		require.NoError(t, err)
		require.Equal(t, 10, int(push.Type))
		require.Equal(t, data, res, conv.MustJSONString(push))
	default:
		assert.Fail(t, "cannot recv push data")
	}
}

func TestRouter_BindModel(t *testing.T) {
	type Foo struct {
		Value int
	}
	addr := fmt.Sprintf("localhost:%d", 1024+rand.Int()%10000)
	s := websocket.NewServer()
	s.Bind("echo", func(ctx context.Context, params interface{}) (interface{}, error) {
		return params.(*Foo), nil
	}).SetModel(&Foo{})
	go func() {
		err := http.ListenAndServe(addr, s)
		require.NoError(t, err)
	}()
	runtime.Gosched()
	c := websocket.NewClient("ws://"+addr, nil)
	ctx := context.Background()
	var res Foo
	err := c.Call(ctx, "echo", Foo{Value: 10}, &res)
	require.NoError(t, err)
	require.Equal(t, 10, res.Value)
	time.Sleep(time.Second)
}
