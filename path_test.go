package wine

import "testing"

func TestIsStaticPath(t *testing.T) {
	if isStaticPath(":a") {
		t.Fail()
	}

	if !isStaticPath("ab") {
		t.Fail()
	}

	if isParamPath("/a") {
		t.Fail()
	}
}

func TestIsParamPath(t *testing.T) {
	if isParamPath(":/a") {
		t.Fail()
	}

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
