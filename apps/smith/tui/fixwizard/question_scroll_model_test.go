package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func longQuestion() doctorfix.Planned {
	var b strings.Builder
	for i := range 60 {
		b.WriteString("line of the explanation number ")
		b.WriteString(strings.Repeat("x", i%7))
		b.WriteString("\n")
	}
	return doctorfix.Planned{Fix: protocol.Fix{Description: "move the builds", Choices: []protocol.FixChoice{
		{Question: b.String() + "Move them?", Options: []protocol.FixChoiceOption{{Label: "move"}, {Label: "leave"}}},
	}}, Ask: true}
}

func TestTheWheelScrollsALongQuestionAndLeavesTheCursor(t *testing.T) {
	w := New(longQuestion(), show).Place(100, 20)
	cursor := w.cursor
	w = w.Wheel(3)
	if w.top != 3 || w.cursor != cursor {
		t.Fatalf("the wheel must scroll the question, not move the cursor: top %d cursor %d", w.top, w.cursor)
	}
}

func TestPageKeysScrollALongQuestionAsItsHintSays(t *testing.T) {
	w := New(longQuestion(), show).Place(100, 20)
	w, _ = press(t, w, "pgdown")
	if w.top == 0 {
		t.Fatal("PgDn scrolls a long question")
	}
	top := w.top
	w, _ = press(t, w, "ctrl+u")
	if w.top >= top {
		t.Fatal("Ctrl+U scrolls it back")
	}
}

// scrollbarOnTheBorder fails unless the lines draw a scrollbar and every cell
// of its track is the last one inside the right border, as in every pane.
func scrollbarOnTheBorder(t *testing.T, lines []string) {
	t.Helper()
	track := []rune(theme.Default().Icons.Track)[0]
	bars := 0
	for _, l := range lines {
		r := []rune(pane.Strip(l))
		for x, c := range r {
			if c != track {
				continue
			}
			if x != len(r)-2 {
				t.Fatalf("the scrollbar is in column %d, not next to the border in column %d: %q", x, len(r)-2, string(r))
			}
			bars++
		}
	}
	if bars == 0 {
		t.Fatalf("the test needs a body that scrolls:\n%s", pane.Strip(strings.Join(lines, "\n")))
	}
}

// The dependency repair is this wizard with its own words, so it is pinned here too.
func TestTheFixAndRepairScrollbarSitsNextToTheBorder(t *testing.T) {
	fix := New(longQuestion(), show).Place(100, 16)
	repair := New(longQuestion(), show).Styled(Style{Name: "repair", NoRetry: true, NoSuggestion: true}).Place(100, 16)
	for _, w := range []Wizard{fix, repair} {
		scrollbarOnTheBorder(t, w.Lines(theme.Default(), 0))
	}
}
