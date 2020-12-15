package respond_test

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

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/gopub/wine/httpvalue"
	"github.com/gopub/wine/internal/respond"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const charsetSuffix = "; charset=utf-8"

func TestHTML(t *testing.T) {
	v := "<html><body>hello</body></html>"
	h := respond.HTML(http.StatusOK, v)
	recorder := httptest.NewRecorder()
	h.Respond(context.Background(), recorder)
	resp := recorder.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httpvalue.HTML+charsetSuffix, resp.Header.Get(httpvalue.ContentType))
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
		j := respond.JSON(http.StatusOK, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, httpvalue.JSON+charsetSuffix, resp.Header.Get(httpvalue.ContentType))
		var result int
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Equal(t, v, result)
	})

	t.Run("String", func(t *testing.T) {
		v := fmt.Sprintf("The number is %d", rand.Int())
		j := respond.JSON(http.StatusOK, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, httpvalue.JSON+charsetSuffix, resp.Header.Get(httpvalue.ContentType))
		var result string
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Equal(t, v, result)
	})

	t.Run("Struct", func(t *testing.T) {
		v := &Item{
			String:  uuid.New().String(),
			Integer: rand.Int(),
			Float:   rand.Float32(),
			Array:   []int{rand.Int(), rand.Int()},
			Time:    time.Now().Add(time.Hour),
		}
		j := respond.JSON(http.StatusCreated, v)
		recorder := httptest.NewRecorder()
		j.Respond(ctx, recorder)
		resp := recorder.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, httpvalue.JSON+charsetSuffix, resp.Header.Get(httpvalue.ContentType))
		var result *Item
		err := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		require.NoError(t, err)
		require.Empty(t, cmp.Diff(v, result))
	})
}
