package wine

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestNewNode(t *testing.T) {
	n := newNode("*")
	assert.Equal(t, wildcardNode, n.t)
	n = newNode("*file")
	assert.Equal(t, wildcardNode, n.t)

	n = newNode("{a}")
	assert.Equal(t, paramNode, n.t)
	assert.Equal(t, "a", n.paramName)
}
