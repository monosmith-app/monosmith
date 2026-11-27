package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/tui/layout"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
)

func TestAnInlineWizardFillsItsAreaWithNoFrameAndItsButtonsAtTheLeft(t *testing.T) {
	area := layout.Rect{X: 10, Y: 5, W: 50, H: 12}
	w := New(safe(), show).Inline(area)
	lines := w.Lines(theme.Default(), 0)
	if len(lines) != area.H {
		t.Fatalf("%d lines, want %d", len(lines), area.H)
	}
	for _, l := range lines {
		if pane.Width(l) != area.W || strings.ContainsAny(pane.Strip(l), "╭╰│") {
			t.Fatalf("%q is not a line of the area", pane.Strip(l))
		}
	}
	if !strings.Contains(pane.Strip(lines[0]), "Fix · review") {
		t.Fatalf("the title is the first line: %q", pane.Strip(lines[0]))
	}
	last := pane.Strip(lines[len(lines)-1])
	if strings.HasPrefix(last, "      ") || !strings.Contains(last, "Apply (Enter)") || strings.Contains(last, "Cancel") {
		t.Fatalf("the buttons start at the left, with no Cancel: %q", last)
	}
	x := area.X + strings.Index(last, "Apply")
	if _, out := w.Click(x, area.Y+area.H-1); !out.Apply {
		t.Fatalf("a click on Apply applies: %+v", out)
	}
	if !w.IsInline() || New(safe(), show).IsInline() {
		t.Fatal("IsInline tells an inline wizard")
	}
}
