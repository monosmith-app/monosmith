package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestDimStripsTheColoursAndKeepsTheCells(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	lipgloss.SetColorProfile(termenv.ANSI256)
	th := Default()
	row := th.ButtonHover(" Home ") + " " + th.Success("ok")
	got := th.Dim(row)
	if got != th.Faint(" Home  ok") || strings.Contains(got, "48;") {
		t.Fatalf("the row is dimmed to the faint style alone: %q", got)
	}
}

func TestDimInThePlainThemeIsTheFaintAttributeOnlyWhereStylesPrint(t *testing.T) {
	defer lipgloss.SetColorProfile(lipgloss.ColorProfile())
	plain := New(Choice{Tokens: Plain, Icons: ASCII})
	lipgloss.SetColorProfile(termenv.ANSI256)
	if got := plain.Dim("a b"); got != "\x1b[2ma b\x1b[m" {
		t.Fatalf("faint attribute: %q", got)
	}
	lipgloss.SetColorProfile(termenv.Ascii)
	if got := plain.Dim("a b"); got != "a b" {
		t.Fatalf("a terminal that prints no styles keeps the row as it is: %q", got)
	}
}
