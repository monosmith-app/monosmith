package guide

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// Bubble Tea draws one last frame after the final answer, when no step is
// left; that frame must render (it once dereferenced a nil step).
func TestTheFrameAfterTheFinalAnswerRenders(t *testing.T) {
	m := keys(menu(), "down", "enter", "down", "down", "enter", "down", "enter", "s", "d", "k", "enter")
	if !m.(Model).Done() {
		t.Fatal("the guide should be done")
	}
	if view := m.View(); view == "" {
		t.Fatal("empty final frame")
	}
	for _, w := range []int{0, 1, 5, 80} {
		lines, _ := m.(Model).Body(theme.Default(), w)
		_ = lines
	}
}

func TestALongEntryWrapsAndEveryLineOfItIsClickable(t *testing.T) {
	long := []Action{{Title: "Adopt this repository", Hint: "smith adopt: bring an existing repository under smith without changing how it builds"}, {Title: "Other"}}
	lines, optionAt := NewModel(long).Body(theme.Default(), 40)
	if len(lines) < 4 {
		t.Fatalf("the hint was cut, not wrapped:\n%s", strings.Join(lines, "\n"))
	}
	text := strings.Join(lines, " ")
	if !strings.Contains(text, "builds") {
		t.Fatalf("the end of the hint is missing:\n%s", strings.Join(lines, "\n"))
	}
	if optionAt[0] != 0 || optionAt[1] != 0 || optionAt[len(lines)-4] != 1 { // the help line below the options wraps onto two lines too
		t.Errorf("rows do not map to their option: %v", optionAt)
	}
}
