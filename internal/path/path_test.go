package path_test

import (
	"testing"

	"github.com/gopub/wine/internal/path"
	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert.Equal(t, "hello", path.Normalize("hello//"))
	assert.Equal(t, "hello/{id}", path.Normalize("hello/{id}/"))
	assert.Equal(t, "hello/{id}", path.Normalize("//hello/{id}/"))
	assert.Empty(t, path.Normalize("//"))
}

func TestIsStaticPath(t *testing.T) {
	assert.Empty(t, path.IsStatic("{a}"))
	assert.NotEmpty(t, path.IsStatic("ab"))
	assert.Empty(t, path.IsParam("/a"))
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
			assert.NotEmpty(t, path.IsParam(v))
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
			assert.Empty(t, path.IsParam(v))
		}
	})
}
