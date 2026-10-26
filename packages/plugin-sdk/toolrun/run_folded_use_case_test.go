package toolrun

import (
	"strings"
	"testing"
)

func TestFoldedKeepsTheOutputOffTheLogWhenItSucceeds(t *testing.T) {
	t.Setenv(VerboseEnv, "")
	log := captureLog(t)
	notes, err := Folded(t.TempDir(), []string{"go", "version"})
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 0 || strings.Contains(log.String(), "go version go") {
		t.Fatalf("notes %q log %q", notes, log.String())
	}
}

func TestFoldedShowsTheOutputWhenItFails(t *testing.T) {
	t.Setenv(VerboseEnv, "")
	log := captureLog(t)
	if _, err := Folded(t.TempDir(), []string{"go", "definitely-not-a-command"}); err == nil {
		t.Fatal("a failing command must fail")
	}
	if !strings.Contains(log.String(), "definitely-not-a-command") {
		t.Fatalf("%q", log.String())
	}
}
