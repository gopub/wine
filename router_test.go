package wine

import "testing"

func TestRouter1(t *testing.T) {
	r := NewRouter()
	r.Get("/hello", func(c Context) bool {
		t.Log("hello")
		return true
	})

	r.Get("/hello/:world", func(c Context) bool {
		t.Log("hello,world")
		return true
	})

	r.Get("/hello/:world2", func(c Context) bool {
		t.Log("hello,world")
		return true
	})

	hs, ps := r.Match("GET", "/hello/123")
	hs[0](nil)
	t.Log(ps)
}
