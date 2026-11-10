package removeproject

import (
	"strings"
	"testing"
)

func TestKeptReleasesLinesNameTheTagsThatStay(t *testing.T) {
	if got := KeptReleasesLines(Removal{Project: "p"}); got != nil {
		t.Fatalf("a project that never released keeps nothing: %v", got)
	}
	got := strings.Join(KeptReleasesLines(Removal{Project: "p", Releases: []string{"p@0.1.0", "p@0.2.0"}}), "\n")
	for _, want := range []string{"its releases stay: tags p@0.1.0, p@0.2.0, GitHub Releases and registry packages are not touched", "did not ask any registry"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	many := []string{"p@1.0.0", "p@1.1.0", "p@1.2.0", "p@1.3.0", "p@1.4.0", "p@1.5.0", "p@1.6.0", "p@1.7.0"}
	if got := KeptReleasesLines(Removal{Releases: many})[0]; !strings.Contains(got, "and 2 more") || strings.Contains(got, "p@1.0.0") {
		t.Errorf("long lists are shortened to the latest: %s", got)
	}
}
