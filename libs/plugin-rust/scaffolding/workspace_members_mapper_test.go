package scaffolding

import "testing"

func TestAddExcludeAppendsOrCreatesTheKey(t *testing.T) {
	for _, c := range []struct{ name, src, want string }{
		{"creates", "[workspace]\nmembers = [\"a\"]\n", "[workspace]\nexclude = [\"tools/x\"]\nmembers = [\"a\"]\n"},
		{"appends", "[workspace]\nexclude = [\"b\"]\nmembers = [\"a\"]\n", "[workspace]\nexclude = [\"b\", \"tools/x\"]\nmembers = [\"a\"]\n"},
		{"again", "[workspace]\nexclude = [\"tools/x\"]\n", "[workspace]\nexclude = [\"tools/x\"]\n"},
	} {
		got, err := AddExclude(c.src, "tools/x")
		if err != nil || got != c.want {
			t.Errorf("%s: %q, %v; want %q", c.name, got, err, c.want)
		}
	}
	if _, err := AddExclude("[package]\nname = \"x\"\n", "tools/x"); err == nil {
		t.Error("a manifest without [workspace] has nothing to edit")
	}
}
