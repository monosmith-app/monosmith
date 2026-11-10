package footer

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/button"
	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var (
	view = []commands.Hint{{Glyph: "b", Label: "build", Key: "b"}, {Glyph: "t", Label: "test", Key: "t"}, {Glyph: "l", Label: "lint", Key: "l"}}
	tail = []commands.Hint{{Glyph: "?", Label: "help", Key: "?"}, {Glyph: "q", Label: "quit", Key: "q"}}
)

func TestTheViewHintsSitLeftAndTheTailRight(t *testing.T) {
	placed := Place(theme.Default(), view, tail, 70)
	if len(placed) != 5 || placed[0].X != 1 {
		t.Fatalf("%+v", placed)
	}
	last := placed[len(placed)-1]
	if last.X+last.W != 69 {
		t.Fatalf("the tail ends at %d", last.X+last.W)
	}
	line := pane.Strip(Render(theme.Default(), placed, -1, 70))
	if pane.Width(line) != 70 || !strings.Contains(line, "build (b)    test (t)    lint (l)") || !strings.HasSuffix(strings.TrimRight(line, " "), "help (?)    quit (q)") {
		t.Fatalf("%q", line)
	}
}

func TestWhenTheyDoNotFitTheViewsLastHintsGoFirstAndQuitIsKept(t *testing.T) {
	placed := Place(theme.Default(), view, tail, 40)
	var labels []string
	for _, p := range placed {
		labels = append(labels, p.Hint.Label)
	}
	if got := strings.Join(labels, ","); got != "build,help,quit" {
		t.Fatalf("%s", got)
	}
	placed = Place(theme.Default(), view, tail, 14)
	if len(placed) != 1 || placed[0].Hint.Label != "quit" {
		t.Fatalf("%+v", placed)
	}
}

func TestAtFindsTheHintUnderAColumn(t *testing.T) {
	placed := Place(theme.Default(), view, tail, 70)
	if h, ok := At(placed, 2); !ok || h.Key != "b" {
		t.Fatalf("%+v %v", h, ok)
	}
	if _, ok := At(placed, 0); ok {
		t.Fatal("column 0 is blank")
	}
}

func TestAHintWithoutAKeyStaysTextAndAKeyedOneIsAButton(t *testing.T) {
	th := theme.Default()
	hints := []commands.Hint{{Glyph: "↑↓", Label: "move"}, {Glyph: "b", Label: "build", Key: "b"}}
	placed := Place(th, hints, nil, 60)
	if placed[0].W != pane.Width("↑↓ move") || placed[1].W != button.Width(th, "build", "b") {
		t.Fatalf("%+v", placed)
	}
	line := pane.Strip(Render(th, placed, -1, 60))
	if !strings.Contains(line, "↑↓ move") || !strings.Contains(line, "build (b)") {
		t.Fatalf("%q", line)
	}
}

func TestTheHintUnderThePointerIsLit(t *testing.T) {
	th := theme.Default()
	placed := Place(th, view, tail, 60)
	if Render(th, placed, -1, 60) == Render(th, placed, placed[0].X, 60) {
		t.Fatal("hover must change the look")
	}
}
