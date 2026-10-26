package venv

import (
	"strings"
	"testing"
)

func TestGitignoreWithAddsTheVenvOnce(t *testing.T) {
	out, changed := GitignoreWith("node_modules/\n")
	if !changed || !strings.Contains(out, "node_modules/") || !strings.Contains(out, ".venv/") {
		t.Fatalf("%v %q", changed, out)
	}
	if _, again := GitignoreWith(out); again {
		t.Fatal("a venv already listed is not listed twice")
	}
	for _, listed := range []string{".venv", "/.venv", "/.venv/", "a\n  .venv/  \n"} {
		if _, changed := GitignoreWith(listed); changed {
			t.Errorf("%q already ignores the venv", listed)
		}
	}
	if out, _ := GitignoreWith(""); out != "# Python\n.venv/\n" {
		t.Fatalf("%q", out)
	}
}
