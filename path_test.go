package wine

import "testing"

func TestIsParamPath(t *testing.T) {
	if !isParamPath(":a,:b") {
		t.Fail()
	}

	if !isParamPath(":a_b") {
		t.Fail()
	}

	if isParamPath("c") {
		t.Fail()
	}
}
