package wine

import (
	"testing"
)

func TestNormalizePath(t *testing.T) {
	if normalizePath("hello//") != "hello" {
		t.FailNow()
	}

	if normalizePath("hello/:id/") != "hello/:id" {
		t.FailNow()
	}

	if normalizePath("//hello/:id/") != "hello/:id" {
		t.FailNow()
	}

	if normalizePath("//") != "" {
		t.FailNow()
	}
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
