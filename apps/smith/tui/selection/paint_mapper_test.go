package selection

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() { lipgloss.SetColorProfile(termenv.ANSI256) }

func TestPaintKeepsEveryWidthAndTheText(t *testing.T) {
	for name, th := range map[string]theme.Theme{
		"dark":  theme.New(theme.Choice{Tokens: theme.Dark, Icons: theme.Unicode}),
		"plain": theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII}),
	} {
		rows := []string{"a\x1b[31m世界\x1b[0mbc", "|alpha beta  |", "xyz"}
		want := []int{ansi.StringWidth(rows[0]), ansi.StringWidth(rows[1]), ansi.StringWidth(rows[2])}
		plain := []string{ansi.Strip(rows[0]), ansi.Strip(rows[1]), ansi.Strip(rows[2])}
		s := State{}.Press(inner, 2, 0).Drag(4, 2)
		got := Paint(th, append([]string(nil), rows...), s)
		for i, line := range got {
			if ansi.StringWidth(line) != want[i] {
				t.Fatalf("%s row %d width %d, want %d: %q", name, i, ansi.StringWidth(line), want[i], line)
			}
			if ansi.Strip(line) != plain[i] {
				t.Fatalf("%s row %d text changed: %q", name, i, line)
			}
		}
		if !strings.Contains(got[1], "\x1b[") || got[1] == rows[1] {
			t.Fatalf("%s: the middle row is highlighted: %q", name, got[1])
		}
	}
}

func TestPaintInReverseVideoWithoutColour(t *testing.T) {
	th := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
	got := Paint(th, []string{"abcdef"}, State{}.Press(inner, 1, 0).Drag(2, 0))
	if !strings.Contains(got[0], "\x1b[7m") {
		t.Fatalf("the plain theme highlights with reverse video: %q", got[0])
	}
}

func TestPaintLeavesTheRowsAloneWhenNothingIsSelected(t *testing.T) {
	th := theme.Default()
	rows := []string{"abc", "def"}
	got := Paint(th, append([]string(nil), rows...), State{}.Press(inner, 1, 0))
	if got[0] != rows[0] || got[1] != rows[1] {
		t.Fatalf("%q", got)
	}
}
