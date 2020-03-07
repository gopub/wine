package path

import (
	"container/list"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	n := NewNode("*", "*")
	assert.Equal(t, WildcardNode, n.typ)
	n = NewNode("*file", "*file")
	assert.Equal(t, WildcardNode, n.typ)

	n = NewNode("{a}", "{a}")
	assert.Equal(t, ParamNode, n.typ)
	assert.Equal(t, "a", n.paramName)
}

func TestNode_Conflict(t *testing.T) {
	hl := list.New()
	hl.PushBack("")
	root := NewNodeList("/hello/world/{param}", hl)

	pair := root.Conflict(NewNodeList("/hello/world/{param}", hl))
	assert.NotEmpty(t, pair)

	pair = root.Conflict(NewNodeList("/hello/{world}", hl))
	assert.Empty(t, pair)

	pair = root.Conflict(NewNodeList("/hello/{world}/{param}", hl))
	assert.Empty(t, pair)

	pair = root.Conflict(NewNodeList("/hello/world/*", hl))
	assert.Empty(t, pair)
}
