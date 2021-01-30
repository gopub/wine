package urlutil_test

import (
	"testing"

	"github.com/gopub/wine/urlutil"
)

func TestJoin(t *testing.T) {
	u := urlutil.Join("http://www.test.com", "hello")
	if u != "http://www.test.com/hello" {
		t.Error(u)
		t.FailNow()
	}
}
