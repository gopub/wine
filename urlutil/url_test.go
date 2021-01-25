package urlutil_test

import (
	"github.com/gopub/wine/urlutil"
	"testing"
)

func TestJoin(t *testing.T) {
	u := urlutil.Join("http://www.test.com", "hello")
	if u != "http://www.test.com/hello" {
		t.Error(u)
		t.FailNow()
	}
}
