package scaffold

import (
	"strings"
	"testing"
)

func TestOnlyALibraryTellsHowAnAppUsesIt(t *testing.T) {
	got := libraryUseHint("lib", "core")
	for _, want := range []string{"smith install -w <app> core", "press + on the app"} {
		if !strings.Contains(got, want) {
			t.Errorf("the hint lacks %q: %q", want, got)
		}
	}
	for _, kind := range []string{"app", "package", ""} {
		if libraryUseHint(kind, "x") != "" {
			t.Errorf("a %q is not imported by an app", kind)
		}
	}
}
