package button

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/muesli/termenv"
)

func init() {
	// Draw escape codes even though the tests have no terminal, so the states can be told apart.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func TestAButtonShowsItsKeyInParentheses(t *testing.T) {
	if got := pane.Strip(Render(theme.Default(), "button", "b", Rest)); got != " button (b) " {
		t.Fatalf("got %q", got)
	}
	if got := pane.Strip(Render(theme.Default(), "Close", "", Rest)); got != " Close " {
		t.Fatalf("without a key: %q", got)
	}
}

func TestEveryStateHasTheSameWidth(t *testing.T) {
	for _, th := range []theme.Theme{theme.Default(), theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})} {
		want := Width(th, "button", "b")
		for _, s := range []State{Rest, Hover, Active, Disabled} {
			if got := pane.Width(Render(th, "button", "b", s)); got != want {
				t.Fatalf("%s state %d: width %d, want %d", th.Tokens.Name, s, got, want)
			}
		}
	}
}

func TestTheThreeStatesAreToldApart(t *testing.T) {
	th := theme.Default()
	rest, hover, active := Render(th, "x", "b", Rest), Render(th, "x", "b", Hover), Render(th, "x", "b", Active)
	if rest == hover || hover == active || rest == active {
		t.Fatalf("states look the same: %q %q %q", rest, hover, active)
	}
}

func TestPlainButtonsUseBracketsAndReverseVideo(t *testing.T) {
	th := theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})
	if got := pane.Strip(Render(th, "button", "b", Rest)); !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "]") {
		t.Fatalf("a plain button has brackets: %q", got)
	}
	if Render(th, "x", "b", Hover) == Render(th, "x", "b", Rest) {
		t.Fatal("hover must differ from rest without colour")
	}
}
