package path

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	if Normalize("hello//") != "hello" {
		t.FailNow()
	}

	if Normalize("hello/{id}/") != "hello/{id}" {
		t.FailNow()
	}

	if Normalize("//hello/{id}/") != "hello/{id}" {
		t.FailNow()
	}

	if Normalize("//") != "" {
		t.FailNow()
	}
}

func TestIsStaticPath(t *testing.T) {
	if IsStatic("{a}") {
		t.Fail()
	}

	if !IsStatic("ab") {
		t.Fail()
	}

	if IsParam("/a") {
		t.Fail()
	}
}

func TestIsParamPath(t *testing.T) {
	if IsParam("{/a}") {
		t.Error("{/a}")
		t.FailNow()
	}

	if !IsParam("{a_b}") {
		t.Error("{a_b}")
		t.FailNow()
	}

	if !IsParam("{_b}") {
		t.Error("{_b}")
		t.FailNow()
	}

	if !IsParam("{_1}") {
		t.Error("{_1}")
		t.FailNow()
	}

	if !IsParam("{a1}") {
		t.Error("{a1}")
		t.FailNow()
	}

	if !IsParam("{a1_}") {
		t.Error("{a1_}")
		t.FailNow()
	}

	if IsParam("c") {
		t.Error("c")
		t.FailNow()
	}

	if IsParam("{_}") {
		t.Error("{_}")
		t.FailNow()
	}

	if IsParam("{__}") {
		t.Error("{__}")
		t.FailNow()
	}

	if IsParam("{a") {
		t.Error("{a")
		t.FailNow()
	}

	if IsParam("{1}") {
		t.Error("{1}")
		t.FailNow()
	}

	if IsParam("{1_a}") {
		t.Error("{1_a}")
		t.FailNow()
	}
}
