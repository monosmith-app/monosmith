package release

import (
	"strings"
	"testing"
)

func TestNothingToReleaseExplainsWhichCommitsCountAndShowsTheFirstReleaseCommand(t *testing.T) {
	text := NothingToRelease()
	for _, want := range []string{"feat:", "fix:", "chore: add app web", "never do", FirstReleaseCommand, "smith does not make it"} {
		if !strings.Contains(text, want) {
			t.Errorf("the explanation lacks %q:\n%s", want, text)
		}
	}
}
