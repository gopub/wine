package wine_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gopub/gox"
	"github.com/gopub/wine"
	"github.com/gopub/wine/mime"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

const charsetSuffix = "; charset=utf-8"

func TestHTML(t *testing.T) {
	v := "<html><body>hello</body></html>"
	h := wine.HTML(http.StatusOK, v)
	recorder := httptest.NewRecorder()
	h.Respond(context.Background(), recorder)
	resp := recorder.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, mime.HTML+charsetSuffix, resp.Header.Get(mime.ContentType))
	result, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)
	require.Equal(t, v, string(result))
}

func TestJSON(t *testing.T) {
	type Item struct {
		String  string
		Integer int
		Float   float32
		Array   []int
		Time    time.Time
	}
	ctx := context.Background()
	t.Run("Int", func(t *testing.T) {
		v := rand.Int()
		j := wine.JSON(http.StatusOK, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, mime.JSON+charsetSuffix, resp.Header.Get(mime.ContentType))
		var result int
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Equal(t, v, result)
	})

	t.Run("String", func(t *testing.T) {
		v := fmt.Sprintf("The number is %d", rand.Int())
		j := wine.JSON(http.StatusOK, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, mime.JSON+charsetSuffix, resp.Header.Get(mime.ContentType))
		var result string
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Equal(t, v, result)
	})

	t.Run("Struct", func(t *testing.T) {
		v := &Item{
			String:  gox.UniqueID64(),
			Integer: rand.Int(),
			Float:   rand.Float32(),
			Array:   []int{rand.Int(), rand.Int()},
			Time:    time.Now().Add(time.Hour),
		}
		j := wine.JSON(http.StatusCreated, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, mime.JSON+charsetSuffix, resp.Header.Get(mime.ContentType))
		var result *Item
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Empty(t, gox.Diff(v, result))
	})
}
