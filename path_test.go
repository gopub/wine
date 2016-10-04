package wine

import (
	"strings"
	"testing"
)

func TestPath1(t *testing.T) {
	t.Log(normalizePath("hello//"))
	t.Log(normalizePath("hello/:id/"))
	t.Log(normalizePath("//hello/:id"))
	t.Log(normalizePath("//"))

	t.Log(len(strings.Split("hello", "/")))
	t.Log(len(strings.Split("hello/:id", "/")))
	t.Log(len(strings.Split("", "/")))
}

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
