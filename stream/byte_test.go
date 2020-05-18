package stream_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/gopub/wine"
	"github.com/gopub/wine/stream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteStream(t *testing.T) {
	packets := [][]byte{
		{0x0A, 0x01, 0x02},
		{},
		[]byte("hello"),
	}
	h := stream.NewByteHandler(func(ctx context.Context, w stream.ByteWriteCloser) {
		for _, p := range packets {
			err := w.Write(p)
			require.NoError(t, err)
		}
		go func() {
			time.Sleep(2 * time.Second)
			err := w.Close()
			require.NoError(t, err)
		}()
	})
	host := "localhost:" + fmt.Sprint(rand.Int()%1000+8000)
	s := wine.NewServer()
	s.Bind(http.MethodGet, "/", h)
	go s.Run(host)

	req, err := http.NewRequest(http.MethodGet, "http://"+host, nil)
	require.NoError(t, err)
	r, err := stream.NewByteReader(http.DefaultClient, req)
	require.NoError(t, err)

	var res [][]byte
	for {
		p, err := r.Read()
		if err != nil {
			r.Close()
			break
		}
		res = append(res, p)
	}
	err = s.Shutdown()
	assert.NoError(t, err)
	require.Equal(t, packets, res)
}
