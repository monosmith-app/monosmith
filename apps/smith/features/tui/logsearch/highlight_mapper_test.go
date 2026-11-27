package logsearch

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/features/tui/theme"
	"github.com/muesli/termenv"
)

func TestAMatchIsLitAndTheCurrentOneStandsOut(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	th := theme.New(theme.Choice{Tokens: theme.Dark, Icons: theme.Unicode})
	row := "a fail and a fail"
	spans := Find([]string{row}, "FAIL")
	lit := Highlight(th, row, spans, spans[1])
	if pane.Strip(lit) != row {
		t.Fatalf("the text stays: %q", pane.Strip(lit))
	}
	if !strings.Contains(lit, th.Accent(th.Bold("fail"))) || !strings.Contains(lit, th.Selected(th.Accent(th.Bold("fail")))) {
		t.Fatalf("both matches are lit, the current one on the selected background: %q", lit)
	}
	if got := Highlight(th, "\x1b[31mred\x1b[0m", nil, Match{}); got != "\x1b[31mred\x1b[0m" {
		t.Fatalf("a row with no match keeps its colour: %q", got)
	}
}
