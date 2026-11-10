package cachewizard

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/monosmith-app/monosmith/apps/smith/cachesetup"
)

func listOf(labels ...string) Model {
	var opts []cachesetup.Option
	for _, l := range labels {
		opts = append(opts, cachesetup.Option{Label: l, Value: l})
	}
	return New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Choose, Prompt: "Which container?", Options: opts}}, time.Now())
}

func press(m Model, key string) Model {
	m, _ = m.Key(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return m
}

func TestALetterJumpsToTheFirstEntryThatStartsWithIt(t *testing.T) {
	m := listOf("applease", "hnsonerror", "kontent-articles", "kontent-extra", "vahan")
	if m = press(m, "k"); m.cursor != 2 {
		t.Fatalf("k went to %d, want the first k entry", m.cursor)
	}
	if m = press(m, "v"); m.cursor != 4 {
		t.Fatalf("v went to %d", m.cursor)
	}
}

func TestLettersTypedQuicklyBuildAWord(t *testing.T) {
	m := listOf("kontent-articles", "kontent-extra", "kind")
	m = press(press(press(m, "k"), "o"), "n")
	if m.cursor != 0 {
		t.Fatalf("kon went to %d", m.cursor)
	}
	if next, ok := m.jumpTo('e', m.typedAt.Add(time.Millisecond)); ok || next.cursor != 0 {
		t.Fatalf("no entry starts with e or kone: the cursor stays, got %d (found %v)", next.cursor, ok)
	}
}

func TestTheSameLetterAgainGoesRoundTheEntriesThatStartWithIt(t *testing.T) {
	m := listOf("kontent-a", "x", "kontent-b")
	m = press(m, "k")
	first := m.cursor
	m.typedAt = time.Time{} // a pause: a new word
	m = press(m, "k")
	if first != 0 || m.cursor != 2 {
		t.Fatalf("k then k went %d then %d, want 0 then 2", first, m.cursor)
	}
}

func TestALetterWithNoEntryKeepsItsOldMeaning(t *testing.T) {
	m := listOf("applease", "banana", "cherry")
	m.cursor = 1
	if m = press(m, "k"); m.cursor != 0 {
		t.Fatalf("k with no k entry is still up: cursor %d", m.cursor)
	}
}

func TestTheWheelRollsTheListAndLeavesTheSelectionWhereItIs(t *testing.T) {
	m := listOf("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	m.cursor = 4
	m = m.Wheel(3)
	if m.cursor != 4 {
		t.Fatalf("the wheel moved the selection to %d", m.cursor)
	}
}
