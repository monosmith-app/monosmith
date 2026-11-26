package rail

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func views() []View {
	return []View{
		{ID: "home", Label: "Home", Key: "1"},
		{ID: "projects", Label: "Projects", Key: "2", Badge: "×1"},
		{ID: "doctor", Label: "Health", Key: "3", Badge: "▲2"},
		{ID: "deps", Label: "Dependencies", Key: "4", Badge: "↑3"},
	}
}

func plainProps(h int) Props {
	return Props{Theme: theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode}), Title: "Views", Views: views(), Active: "projects", Width: 24, Height: h}
}

func plain(lines []string) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = pane.Strip(l)
	}
	return out
}

func TestTheRailIsAFrameWithOneViewPerRowAndABlankRowBetween(t *testing.T) {
	r := Render(plainProps(14))
	lines := plain(r.Lines)
	if len(lines) != 14 {
		t.Fatalf("rows: %d", len(lines))
	}
	for i, l := range lines {
		if pane.Width(l) != 24 {
			t.Fatalf("row %d is %d cells: %q", i, pane.Width(l), l)
		}
	}
	want := []string{
		"╭─ Views ──────────────╮",
		"│                      │",
		"│   Home             1 │",
		"│                      │",
		"│ ▶ Projects     ×1  2 │",
		"│                      │",
		"│   Health       ▲2  3 │",
		"│                      │",
		"│   Dependencies ↑3  4 │",
	}
	for i, w := range want {
		if lines[i] != w {
			t.Fatalf("row %d = %q, want %q\n%s", i, lines[i], w, strings.Join(lines, "\n"))
		}
	}
	if !strings.HasPrefix(lines[13], "╰") {
		t.Fatalf("the frame closes on the last row: %q", lines[13])
	}
}

func TestEveryRowOfAViewFindsIt(t *testing.T) {
	r := Render(plainProps(14))
	for y, id := range map[int]string{2: "home", 4: "projects", 6: "doctor", 8: "deps"} {
		if got := r.At(y); got != id {
			t.Errorf("row %d = %q, want %q", y, got, id)
		}
	}
	for _, y := range []int{0, 1, 3, 9, 13} {
		if got := r.At(y); got != "" {
			t.Errorf("row %d holds no view, got %q", y, got)
		}
	}
}

func TestAShortRailDropsTheBlankRows(t *testing.T) {
	r := Render(plainProps(7))
	lines := plain(r.Lines)
	if !strings.Contains(lines[1], "Home") || !strings.Contains(lines[2], "Projects") || r.At(4) != "deps" {
		t.Fatalf("%s", strings.Join(lines, "\n"))
	}
}

func TestTheOpenViewIsOnTheHighlightAndTheHoveredOneOnTheHover(t *testing.T) {
	p := plainProps(14)
	p.Theme = theme.Default()
	r := Render(p)
	if !strings.Contains(r.Lines[4], p.Theme.Accent("▶ ")) {
		t.Fatalf("the open view's mark is in the accent: %q", r.Lines[4])
	}
	if r.Lines[4] == Render(Props{Theme: p.Theme, Title: p.Title, Views: p.Views, Width: 24, Height: 14}).Lines[4] {
		t.Fatal("the open view's row must be drawn differently from a resting one")
	}
	p.Hover = "home"
	if Render(p).Lines[2] == r.Lines[2] {
		t.Fatal("the hovered view's row must be drawn differently from a resting one")
	}
}

func TestTinySizesDoNotPanic(t *testing.T) {
	for w := 0; w < 6; w++ {
		for h := 0; h < 4; h++ {
			p := plainProps(h)
			p.Width = w
			Render(p)
		}
	}
}
