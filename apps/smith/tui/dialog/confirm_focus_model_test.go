package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func releaseDialog() Confirm {
	return NewConfirm("Release now", "It cannot be undone.", []string{"release", "--yes"}).WithRunLabel("Release anyway").FocusCancel().Place(100, 30)
}

func TestACancelFocusedDialogCancelsOnEnterAndConfirmsOnY(t *testing.T) {
	if _, out := releaseDialog().Key("enter"); !out.Done || out.Confirmed {
		t.Fatalf("Enter: %+v", out)
	}
	if _, out := releaseDialog().Key("esc"); !out.Done || out.Confirmed {
		t.Fatalf("Esc: %+v", out)
	}
	if _, out := releaseDialog().Key("y"); !out.Done || !out.Confirmed {
		t.Fatalf("y: %+v", out)
	}
	c, _ := releaseDialog().Key("tab")
	if _, out := c.Key("enter"); !out.Done || !out.Confirmed {
		t.Fatalf("Tab then Enter: %+v", out)
	}
}

func TestTheRunButtonCarriesTheLabelAndTheKeyThatConfirms(t *testing.T) {
	out := strings.Join(releaseDialog().Lines(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})), "\n")
	if !strings.Contains(out, "Release anyway (y)") || !strings.Contains(out, "Cancel (esc)") {
		t.Fatalf("buttons:\n%s", out)
	}
}

func TestADialogWithItsOwnLineDrawsIt(t *testing.T) {
	c := NewConfirm("First", "One commit.", []string{"git", "commit"}).WithLine(`git commit --allow-empty -m "feat: first version"`).Place(100, 30)
	out := strings.Join(c.Lines(theme.New(theme.Choice{Tokens: theme.Plain, Icons: theme.Unicode})), "\n")
	if !strings.Contains(out, `$ git commit --allow-empty -m "feat: first version"`) || strings.Contains(out, "$ smith") {
		t.Fatalf("line:\n%s", out)
	}
}
