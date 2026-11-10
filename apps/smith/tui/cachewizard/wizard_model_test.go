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

func TestSecretPressMapsEnterEscAndLetters(t *testing.T) {
	for name, tc := range map[string]struct {
		msg  tea.KeyMsg
		want rune
	}{
		"enter":  {tea.KeyMsg{Type: tea.KeyEnter}, '\r'},
		"esc":    {tea.KeyMsg{Type: tea.KeyEsc}, 27},
		"ctrl+c": {tea.KeyMsg{Type: tea.KeyCtrlC}, 3},
		"letter": {tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, 'h'},
		"other":  {tea.KeyMsg{Type: tea.KeyTab}, 1},
	} {
		if got := secretPress(tc.msg).Press; got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}

// screens are one screen of each kind that asks, with a way back or without.
func screens(canBack bool) map[string]Model {
	opts := []cachesetup.Option{{Label: "Azure", Value: "azure"}, {Label: "S3", Value: "s3"}}
	return map[string]Model{
		"choose":  New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Choose, Prompt: "Kind?", Options: opts, CanBack: canBack}}, time.Now()),
		"text":    New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Text, Prompt: "Address", CanBack: canBack}}, time.Now()),
		"confirm": New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Confirm, Confirmation: cachesetup.Confirmation{Prompt: "Write it?", Yes: "Yes", No: "No"}, CanBack: canBack}}, time.Now()),
		"page":    New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Page, Prompt: "Steps", Lines: []string{"one"}, CanBack: canBack}}, time.Now()),
	}
}

func keyOf(name string) tea.KeyMsg {
	switch name {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(name)}
}

func TestEscLeftAndBGoBackWhenAQuestionIsBefore(t *testing.T) {
	for kind, m := range screens(true) {
		keys := []string{"esc", "left", "b"}
		if kind == "text" {
			keys = keys[:2] // a letter is typed in a text field
		}
		for _, k := range keys {
			next, out := m.Key(keyOf(k))
			if out.Reply == nil || !out.Reply.Back || out.Cancel {
				t.Errorf("%s %s: %+v, want a Back reply", kind, k, out)
			}
			if next.Kind() != Idle {
				t.Errorf("%s %s: the screen waits for the engine after Back", kind, k)
			}
		}
	}
}

func TestEscClosesOnTheFirstQuestionAndLeftAndBDoNothing(t *testing.T) {
	for kind, m := range screens(false) {
		if _, out := m.Key(keyOf("esc")); !out.Cancel || out.Reply != nil {
			t.Errorf("%s: esc on the first question closes: %+v", kind, out)
		}
		for _, k := range []string{"left", "b"} {
			if kind == "text" && k == "b" {
				continue
			}
			if _, out := m.Key(keyOf(k)); out.Cancel || out.Reply != nil {
				t.Errorf("%s %s: no question before, nothing happens: %+v", kind, k, out)
			}
		}
	}
}

func TestCtrlCClosesEveryScreenEvenWithAWayBack(t *testing.T) {
	for kind, m := range screens(true) {
		if _, out := m.Key(keyOf("ctrl+c")); !out.Cancel {
			t.Errorf("%s: ctrl+c closes the wizard: %+v", kind, out)
		}
	}
}

func TestInATextFieldLeftGoesBackOnlyWhenNothingIsTyped(t *testing.T) {
	m := screens(true)["text"]
	m = press(m, "x")
	if _, out := m.Key(keyOf("left")); out.Reply != nil {
		t.Fatal("Left with a typed line keeps the line")
	}
	m.input = nil
	if _, out := m.Key(keyOf("left")); out.Reply == nil || !out.Reply.Back {
		t.Fatal("Left with an empty line goes back")
	}
}

func TestBJumpsToAnEntryThatStartsWithBAndOtherwiseGoesBack(t *testing.T) {
	m := listOf("applease", "backup")
	m.scr.CanBack = true
	next, out := m.Key(keyOf("b"))
	if out.Reply != nil || next.cursor != 1 {
		t.Fatalf("b jumps to backup: cursor %d %+v", next.cursor, out)
	}
	m = listOf("applease", "cherry")
	m.scr.CanBack = true
	if _, out := m.Key(keyOf("b")); out.Reply == nil || !out.Reply.Back {
		t.Fatalf("b with no b entry goes back: %+v", out)
	}
}

func TestTheEarlierAnswerIsMarkedWhenAScreenComesAgain(t *testing.T) {
	no := cachesetup.Confirmation{Prompt: "Write it?", Yes: "Yes", No: "No", OfferCommit: true, Earlier: &cachesetup.ConfirmAnswer{Yes: false}}
	m := New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Confirm, Confirmation: no}}, time.Now())
	if _, out := m.Key(tea.KeyMsg{Type: tea.KeyEnter}); out.Reply == nil || out.Reply.Yes {
		t.Fatalf("Enter takes the marked answer, which is no: %+v", out)
	}
	yes := cachesetup.Confirmation{Prompt: "Write it?", Yes: "Yes", No: "No", OfferCommit: true, Earlier: &cachesetup.ConfirmAnswer{Yes: true, Commit: true}}
	m = New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Confirm, Confirmation: yes}}, time.Now())
	if _, out := m.Key(tea.KeyMsg{Type: tea.KeyEnter}); out.Reply == nil || !out.Reply.Yes || !out.Reply.Commit {
		t.Fatalf("Enter takes yes with the commit box ticked as before: %+v", out)
	}
	m = New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Text, Prompt: "Address", Typed: "https://a.dev"}}, time.Now())
	if _, out := m.Key(tea.KeyMsg{Type: tea.KeyEnter}); out.Reply == nil || out.Reply.Value != "https://a.dev" {
		t.Fatalf("the earlier line is typed again: %+v", out)
	}
}

func TestTheSecretScreenHasNoBackKeysAndHoldsNoSecret(t *testing.T) {
	m := New(time.Now()).Handle(AskEvent{Screen: Screen{Kind: Secret, Noun: "Key", CanBack: true, Buttons: []Button{{'q', "Done"}}}}, time.Now())
	for _, k := range []string{"esc", "left", "b", "ctrl+c"} {
		if _, out := m.Key(keyOf(k)); out.Reply != nil || out.Cancel || out.Press == 0 {
			t.Errorf("%s on the secret screen goes to the session, never Back: %+v", k, out)
		}
	}
}
