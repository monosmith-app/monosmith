package runpage

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
)

func TestTheEmptyStateSaysHowToStartARun(t *testing.T) {
	lines := Empty(plain, Keys{Build: "b", Test: "t", Palette: "Ctrl+K", Projects: "2"}, 60)
	text := pane.Strip(strings.Join(lines, "\n"))
	for _, want := range []string{"Nothing ran yet", "b  build", "t  test", "Ctrl+K  run any task", "2  open Projects"} {
		if !strings.Contains(text, want) {
			t.Fatalf("the empty state lacks %q:\n%s", want, text)
		}
	}
	for _, l := range lines {
		if pane.Width(l) > 60 {
			t.Fatalf("a line is wider than the pane: %q", l)
		}
	}
}
