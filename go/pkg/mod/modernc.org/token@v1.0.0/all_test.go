package token // import "modernc.org/token"

import (
	"testing"
)

func Test(t *testing.T) {
	f := NewFile("foo", 11)
	f.AddLine(10)
	for i, v := range []struct {
		int
		string
	}{
		{0, "foo:1:1"},
		{9, "foo:1:10"},
		{10, "foo:2:1"},
		{11, "foo:2:2"},
	} {
		if g, e := f.Position(f.Pos(v.int)).String(), v.string; g != e {
			t.Errorf("%v: %q %q", i, g, e)
		}
	}
}
