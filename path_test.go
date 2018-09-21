package wine

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	if normalizePath("hello//") != "hello" {
		t.FailNow()
	}

	if normalizePath("hello/{id}/") != "hello/{id}" {
		t.FailNow()
	}

	if normalizePath("//hello/{id}/") != "hello/{id}" {
		t.FailNow()
	}

	if normalizePath("//") != "" {
		t.FailNow()
	}
}

func TestIsStaticPath(t *testing.T) {
	if isStaticPath("{a}") {
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
	if isParamPath("{/a}") {
		t.Error("{/a}")
		t.FailNow()
	}

	if !isParamPath("{a_b}") {
		t.Error("{a_b}")
		t.FailNow()
	}

	if !isParamPath("{_b}") {
		t.Error("{_b}")
		t.FailNow()
	}

	if !isParamPath("{_1}") {
		t.Error("{_1}")
		t.FailNow()
	}

	if !isParamPath("{a1}") {
		t.Error("{a1}")
		t.FailNow()
	}

	if !isParamPath("{a1_}") {
		t.Error("{a1_}")
		t.FailNow()
	}

	if isParamPath("c") {
		t.Error("c")
		t.FailNow()
	}

	if isParamPath("{_}") {
		t.Error("{_}")
		t.FailNow()
	}

	if isParamPath("{__}") {
		t.Error("{__}")
		t.FailNow()
	}

	if isParamPath("{a") {
		t.Error("{a")
		t.FailNow()
	}

	if isParamPath("{1}") {
		t.Error("{1}")
		t.FailNow()
	}

	if isParamPath("{1_a}") {
		t.Error("{1_a}")
		t.FailNow()
	}
}
