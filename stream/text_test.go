package stream_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	"github.com/gopub/wine"
	"github.com/gopub/wine/stream"
	"github.com/stretchr/testify/require"
)

func TestTextStream(t *testing.T) {
	packets := []string{
		"测试",
		"Test",
		"",
	}
	h := stream.NewTextHandler(func(ctx context.Context, w stream.TextWriteCloser) {
		for _, s := range packets {
			err := w.Write(s)
			require.NoError(t, err)
		}
		err := w.Close()
		require.NoError(t, err)
	})
	host := "localhost:" + fmt.Sprint(rand.Int()%1000+8000)
	s := wine.NewServer()
	s.Bind(http.MethodGet, "/", h)
	go s.Run(host)

	req, err := http.NewRequest(http.MethodGet, "http://"+host, nil)
	require.NoError(t, err)
	r, err := stream.NewTextReader(http.DefaultClient, req)
	require.NoError(t, err)

	var res []string
	for {
		s, err := r.Read()
		if err != nil {
			r.Close()
			break
		}
		res = append(res, s)
	}
	s.Shutdown()
	require.Equal(t, packets, res)
}
