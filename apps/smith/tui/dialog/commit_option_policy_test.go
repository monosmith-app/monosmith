package dialog

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheCommitBoxStartsUntickedAndAddsTheFlagWhenTicked(t *testing.T) {
	c := NewConfirm("Add a project", "Adds it.", []string{"add", "app", "--yes"}).Place(100, 30)
	if !c.Offer || c.Commit {
		t.Fatalf("offer %v commit %v", c.Offer, c.Commit)
	}
	text := strings.Join(c.Lines(theme.Default()), "\n")
	if !strings.Contains(text, "[ ] "+CommitLabel) || strings.Contains(text, "--commit") {
		t.Fatalf("unticked dialog:\n%s", text)
	}
	if got := strings.Join(c.Final(), " "); got != "add app --yes" {
		t.Fatalf("final %q", got)
	}
	c, out := c.Key("c")
	if out.Done || !c.Commit {
		t.Fatalf("c must tick the box: %+v %+v", c, out)
	}
	text = strings.Join(c.Lines(theme.Default()), "\n")
	if !strings.Contains(text, "[x] "+CommitLabel) || !strings.Contains(text, "--commit") {
		t.Fatalf("ticked dialog:\n%s", text)
	}
	if got := strings.Join(c.Final(), " "); got != "add app --yes --commit" {
		t.Fatalf("final %q", got)
	}
}

func TestAClickOnTheCommitBoxTicksIt(t *testing.T) {
	c := NewConfirm("Remove", "Removes it.", []string{"remove", "app", "--yes"}).Place(100, 30)
	r := c.Rect()
	c, out := c.Click(r.X+4, r.Y+1+c.commitRow())
	if out.Done || !c.Commit {
		t.Fatalf("click did not tick: %+v %+v", c, out)
	}
}

func TestNoCommitBoxForCommandsThatCannotCommit(t *testing.T) {
	c := release()
	if c.Offer {
		t.Fatal("release must not offer the commit box")
	}
	if _, out := c.Key("c"); out.Done {
		t.Fatal("c must do nothing")
	}
}

func TestCommittableCommands(t *testing.T) {
	for argv, want := range map[string]bool{
		"adopt": true, "new app": true, "add x": true, "remove x": true, "convert": true, "upgrade": true,
		"release": false, "undo --yes": false, "add x --commit": false, "": false,
	} {
		if got := Committable(strings.Fields(argv)); got != want {
			t.Errorf("Committable(%q) = %v, want %v", argv, got, want)
		}
	}
}
