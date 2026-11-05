package pathfinder

import "testing"

func TestSplitValue(t *testing.T) {
	for _, c := range []struct{ in, dir, prefix string }{
		{"", "", ""},
		{"dist/", "dist/", ""},
		{"packages/we", "packages/", "we"},
		{`packages\web\sr`, "packages/web/", "sr"},
		{"ci.dev", "", "ci.dev"},
	} {
		if d, p := splitValue(c.in); d != c.dir || p != c.prefix {
			t.Errorf("splitValue(%q) = %q, %q; want %q, %q", c.in, d, p, c.dir, c.prefix)
		}
	}
}

func TestParentOf(t *testing.T) {
	for in, want := range map[string]string{"a/b/": "a/", "a/": "", "": "../", "../": "../../", "a/../": "a/../../"} {
		if got := parentOf(in); got != want {
			t.Errorf("parentOf(%q) = %q; want %q", in, got, want)
		}
	}
}
