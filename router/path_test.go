package router_test

import (
	"testing"

	"github.com/gopub/wine/router"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert.Equal(t, "hello", router.Normalize("hello//"))
	assert.Equal(t, "hello/{id}", router.Normalize("hello/{id}/"))
	assert.Equal(t, "hello/{id}", router.Normalize("//hello/{id}/"))
	assert.Empty(t, router.Normalize("//"))
}

func TestIsStaticPath(t *testing.T) {
	assert.Empty(t, router.IsStatic("{a}"))
	assert.NotEmpty(t, router.IsStatic("ab"))
	assert.Empty(t, router.IsParam("/a"))
}

func TestIsParamPath(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		trueCases := []string{
			"{a_b}",
			"{_b}",
			"{_1}",
			"{a1}",
			"{a1_}",
		}
		for _, v := range trueCases {
			assert.NotEmpty(t, router.IsParam(v))
		}
	})
	t.Run("false", func(t *testing.T) {
		falseCases := []string{
			"{/a}",
			"c",
			"{_}",
			"{__}",
			"{a",
			"{1}",
			"{1_a}",
		}
		for _, v := range falseCases {
			assert.Empty(t, router.IsParam(v))
		}
	})
}
