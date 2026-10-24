package footer

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/commands"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

var (
	view = []commands.Hint{{Glyph: "b", Label: "build", Key: "b"}, {Glyph: "t", Label: "test", Key: "t"}, {Glyph: "l", Label: "lint", Key: "l"}}
	tail = []commands.Hint{{Glyph: "?", Label: "help", Key: "?"}, {Glyph: "q", Label: "quit", Key: "q"}}
)

func TestTheViewHintsSitLeftAndTheTailRight(t *testing.T) {
	placed := Place(view, tail, 60)
	if len(placed) != 5 || placed[0].X != 1 {
		t.Fatalf("%+v", placed)
	}
	last := placed[len(placed)-1]
	if last.X+last.W != 59 {
		t.Fatalf("the tail ends at %d", last.X+last.W)
	}
	line := pane.Strip(Render(theme.Default(), placed, -1, 60))
	if pane.Width(line) != 60 || !strings.Contains(line, "b build  t test  l lint") || !strings.HasSuffix(strings.TrimRight(line, " "), "? help  q quit") {
		t.Fatalf("%q", line)
	}
}

func TestWhenTheyDoNotFitTheViewsLastHintsGoFirstAndQuitIsKept(t *testing.T) {
	placed := Place(view, tail, 30)
	var labels []string
	for _, p := range placed {
		labels = append(labels, p.Hint.Label)
	}
	if got := strings.Join(labels, ","); got != "build,help,quit" {
		t.Fatalf("%s", got)
	}
	placed = Place(view, tail, 14)
	if len(placed) != 1 || placed[0].Hint.Label != "quit" {
		t.Fatalf("%+v", placed)
	}
}

func TestAtFindsTheHintUnderAColumn(t *testing.T) {
	placed := Place(view, tail, 60)
	if h, ok := At(placed, 2); !ok || h.Key != "b" {
		t.Fatalf("%+v %v", h, ok)
	}
	if _, ok := At(placed, 0); ok {
		t.Fatal("column 0 is blank")
	}
}
