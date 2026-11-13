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

func TestWhenTheyDoNotFitTheGlobalHintsGoFirstAndTheViewKeepsItsKeys(t *testing.T) {
	labels := func(placed []Placed) string {
		var out []string
		for _, p := range placed {
			out = append(out, p.Hint.Label)
		}
		return strings.Join(out, ",")
	}
	th := theme.Default()
	global := append([]commands.Hint{{Glyph: "m", Label: "mouse", Key: "m"}, {Glyph: ":", Label: "palette", Key: ":"}}, tail...)
	fitting := 2 + total(th, view) + gap + total(th, tail)
	for _, c := range []struct {
		w    int
		want string
	}{
		{fitting + gap + total(th, global[:2]), "build,test,lint,mouse,palette,help,quit"},
		{fitting, "build,test,lint,help,quit"},
		{fitting - 1, "build,test,help,quit"},
		{2 + total(th, view[:1]) + gap + total(th, tail), "build,help,quit"},
		{2 + total(th, view[:1]) + gap + total(th, tail) - 1, "help,quit"},
		{2 + total(th, tail[1:]) + gap, "quit"},
	} {
		if got := labels(Place(th, view, global, c.w)); got != c.want {
			t.Errorf("%d columns: %s, want %s", c.w, got, c.want)
		}
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
