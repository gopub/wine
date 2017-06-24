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

	n = newNode(":a,:b")
	if n.t != _ParamNode {
		t.FailNow()
	}

	if n.paramNames[0] != "a" {
		t.FailNow()
	}

	if n.paramNames[1] != "b" {
		t.FailNow()
	}

}

func TestNewNodeList(t *testing.T) {

}
