package dependencies

import (
	"strings"
	"testing"
)

func TestATypoOfAProjectNameSaysItMatchesNothingAndOffersTheClosest(t *testing.T) {
	got := UnknownNameHint([]string{"api-py", "core-py", "web"}, []string{"core-pyy", "left-pad"})
	if !strings.Contains(got, "core-pyy matches no project of this workspace and the registry does not have it") || !strings.Contains(got, "core-py") {
		t.Errorf("got %q", got)
	}
	if strings.Contains(got, "left-pad") {
		t.Errorf("a name far from every project gets no suggestion: %q", got)
	}
	if got := UnknownNameHint([]string{"core-py"}, []string{"core-py"}); got != "" {
		t.Errorf("a real project is not a typo: %q", got)
	}
}
