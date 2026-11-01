package fixwizard

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/tui/pane"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func fieldWizard() Wizard {
	p := doctorfix.Planned{Fix: protocol.Fix{Description: "where", Choices: []protocol.FixChoice{{
		Question: "Which?",
		Options:  []protocol.FixChoiceOption{{Label: "apps/{name}"}, {Label: "libs/{name}"}},
	}}}}
	return New(p, nil).Styled(Style{Input: "Name"}).WithText("web").Place(100, 30)
}

func typeKeys(w Wizard, keys ...string) Wizard {
	for _, k := range keys {
		w, _ = w.Key(k)
	}
	return w
}

func TestTheTextFieldIsEditedWithEAndShownInTheLabels(t *testing.T) {
	w := typeKeys(fieldWizard(), "e", "backspace", "x", "-", "1")
	if w.Text() != "wex-1" || !w.Editing() {
		t.Fatalf("text %q editing %v", w.Text(), w.Editing())
	}
	screen := pane.Strip(strings.Join(w.Lines(plain(), 0), "\n"))
	for _, want := range []string{"1) apps/wex-1", "Name: wex-1", "Type the name"} {
		if !strings.Contains(screen, want) {
			t.Errorf("screen lacks %q:\n%s", want, screen)
		}
	}
}

func TestWhileEditingKeysAreTextAndEscStopsOnlyTheEditing(t *testing.T) {
	w := typeKeys(fieldWizard(), "e", "q", "j", "esc")
	if w.Text() != "webqj" || w.Editing() || w.Phase() != Asking {
		t.Fatalf("text %q editing %v phase %v", w.Text(), w.Editing(), w.Phase())
	}
	if _, out := w.Key("esc"); !out.Done {
		t.Fatal("Esc outside the editing must cancel")
	}
}

func TestTheTextIsKeptWhenAnOptionIsChosenAndIsLimited(t *testing.T) {
	w := fieldWizard().WithText(strings.Repeat("a", maxText))
	w = typeKeys(w, "e", "b")
	if len(w.Text()) != maxText {
		t.Fatalf("text has %d characters, want at most %d", len(w.Text()), maxText)
	}
	w = typeKeys(w, "enter", "2")
	if w.Text() != strings.Repeat("a", maxText) || w.Phase() != Reviewing {
		t.Fatalf("text %q phase %v", w.Text(), w.Phase())
	}
}
