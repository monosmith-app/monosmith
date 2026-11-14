package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestAnotherBoxReplacesTheCommitBoxAndNeverAddsACommit(t *testing.T) {
	c := NewConfirm("smith 2.0.0 is out", "Update smith now?", []string{"update"}).WithBox("Don't remind me", "d").Place(100, 30)
	text := strings.Join(c.Lines(theme.Default()), "\n")
	if !strings.Contains(text, "Don't remind me  (D)") || strings.Contains(text, "Commit this change") {
		t.Fatalf("box line wrong in\n%s", text)
	}
	c, _ = c.Key("d")
	if !c.BoxTicked() {
		t.Fatal("the key did not tick the box")
	}
	if got := strings.Join(c.Final(), " "); got != "update" {
		t.Errorf("Final = %q, a commit flag was added", got)
	}
	if c, _ = c.Key("c"); !c.BoxTicked() {
		t.Error("the commit key changed the other box")
	}
	c, _ = c.Key("D")
	if c.BoxTicked() {
		t.Error("the key did not untick the box")
	}
}
