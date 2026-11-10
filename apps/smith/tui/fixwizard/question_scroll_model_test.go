package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
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
