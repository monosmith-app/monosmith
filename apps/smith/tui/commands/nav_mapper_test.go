package commands

import "testing"

func TestNavMapsEveryKeyTheViewsUse(t *testing.T) {
	cases := map[string]NavAction{
		"up": NavUp, "k": NavUp, "down": NavDown, "j": NavDown,
		"pgup": NavPageUp, "pgdown": NavPageDown, "home": NavHome, "end": NavEnd,
		"enter": NavConfirm, " ": NavConfirm, "space": NavConfirm,
		"esc": NavCancel, "ctrl+c": NavCancel, "q": NavCancel,
		"tab": NavNext, "shift+tab": NavPrev, "y": NavYes, "n": NavNo,
		"x": NavNone, "": NavNone, "left": NavNone,
	}
	for key, want := range cases {
		if got := Nav(key); got != want {
			t.Errorf("Nav(%q) = %d, want %d", key, got, want)
		}
	}
}

func TestMovesCoversTheCursorActionsOnly(t *testing.T) {
	for a := NavNone; a <= NavNo; a++ {
		if want := a >= NavUp && a <= NavEnd; a.Moves() != want {
			t.Errorf("Moves(%d) = %v", a, a.Moves())
		}
	}
}

func TestNavNamedLeavesTheLettersToATextField(t *testing.T) {
	for _, key := range []string{"k", "j", "q", "y", "n", " ", "space", "x"} {
		if got := NavNamed(key); got != NavNone {
			t.Errorf("NavNamed(%q) = %d, want NavNone", key, got)
		}
	}
	if NavNamed("up") != NavUp || NavNamed("pgdown") != NavPageDown || NavNamed("enter") != NavConfirm {
		t.Error("the named keys must still map")
	}
}
