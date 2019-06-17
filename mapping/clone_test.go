package mapping_test

import (
	"testing"

	"github.com/gopub/gox"
	"github.com/gopub/wine/mapping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	t.Run("CloneToPtr", func(t *testing.T) {
		var dst *gox.PhoneNumber
		src := &gox.PhoneNumber{
			CountryCode:    86,
			NationalNumber: 13800000001,
		}
		err := mapping.Clone(&dst, src)
		require.NoError(t, err)
		assert.Equal(t, src, dst)
	})

}
