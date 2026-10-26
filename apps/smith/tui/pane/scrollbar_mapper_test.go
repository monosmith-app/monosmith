package pane

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func asciiTheme() theme.Theme {
	return theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.ASCII})
}

func bar(rows []string) string {
	var b strings.Builder
	for _, r := range rows {
		b.WriteString(Strip(r)[len(Strip(r))-1:])
	}
	return b.String()
}

func TestAListThatFitsHasNoScrollbar(t *testing.T) {
	rows := []string{"a   ", "b   "}
	got := Scrollbar(asciiTheme(), rows, 4, 0, 2)
	if got[0] != "a   " || got[1] != "b   " {
		t.Fatalf("%q", got)
	}
}

func TestTheThumbIsTheLengthOfWhatShowsAndMovesWithTheTop(t *testing.T) {
	rows := make([]string, 10)
	for i := range rows {
		rows[i] = strings.Repeat("x", 6)
	}
	th := asciiTheme()
	if got := bar(Scrollbar(th, rows, 6, 0, 40)); got != "##||||||||" {
		t.Fatalf("at the top: %q", got)
	}
	if got := bar(Scrollbar(th, rows, 6, 30, 40)); got != "||||||||##" {
		t.Fatalf("at the end: %q", got)
	}
	mid := bar(Scrollbar(th, rows, 6, 15, 40))
	if strings.Count(mid, "#") != 2 || strings.HasPrefix(mid, "#") || strings.HasSuffix(mid, "#") {
		t.Fatalf("in the middle: %q", mid)
	}
	for _, r := range Scrollbar(th, rows, 6, 0, 40) {
		if Width(r) != 6 {
			t.Fatalf("a row is %d cells: %q", Width(r), r)
		}
	}
}

func TestTheThumbIsNeverShorterThanOneRowAndTopIsClamped(t *testing.T) {
	rows := []string{"xxxx", "xxxx", "xxxx"}
	th := asciiTheme()
	if got := bar(Scrollbar(th, rows, 4, 0, 3000)); got != "#||" {
		t.Fatalf("%q", got)
	}
	if got := bar(Scrollbar(th, rows, 4, 99999, 3000)); got != "||#" {
		t.Fatalf("a top past the end: %q", got)
	}
}
