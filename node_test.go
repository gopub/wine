package wine

import "testing"

func TestNewNode(t *testing.T) {
	n := newNode("*")
	if n.t != _WildcardNode {
		t.FailNow()
	}

	n = newNode("*file")
	if n.t != _WildcardNode {
		t.FailNow()
	}

	n = newNode("{a}")
	if n.t != _ParamNode {
		t.FailNow()
	}

	if n.paramName != "a" {
		t.FailNow()
	}
}

func TestNewNodeList(t *testing.T) {

}
