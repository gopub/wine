package stream_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/stream"
	"github.com/stretchr/testify/require"
)

type jsonPacket struct {
	Name     string
	Score    int
	Birthday time.Time
}

func TestJSONStream(t *testing.T) {
	packets := []jsonPacket{
		{
			"tom",
			80,
			time.Now().Add(-time.Hour * 29999),
		},
		{
			"jim john",
			70,
			time.Now().Add(-time.Hour * 3),
		},
		{
			"jim johnson",
			99,
			time.Now().Add(-time.Hour),
		},
	}
	h := stream.NewJSONHandler(func(ctx context.Context, w stream.JSONWriteCloser) {
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
	r, err := stream.NewJSONReader(http.DefaultClient, req)
	require.NoError(t, err)

	var res []jsonPacket
	for {
		var p jsonPacket
		err := r.Read(&p)
		if err != nil {
			r.Close()
			break
		}
		res = append(res, p)
	}
	err = s.Shutdown()
	assert.NoError(t, err)
	require.Empty(t, gox.Diff(packets, res))
}
