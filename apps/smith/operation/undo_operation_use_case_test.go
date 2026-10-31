package operation

import (
	"os"
	"strings"
	"testing"
)

func TestUndoRevertsTheLastOperationWithANewCommit(t *testing.T) {
	g := scratch(t, true)
	before, _ := g.Run("rev-parse", "HEAD")
	execute(t, "billing")
	c, err := Prepare(g.Dir)
	if err != nil {
		t.Fatal(err)
	}
	plan := c.Plan()
	if plan.Empty() || !strings.Contains(plan.Title, "add lib billing") {
		t.Fatalf("plan: %+v", plan)
	}
	if _, err := Undo(g.Dir, c); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat("libs"); err == nil {
		t.Fatal("libs must be gone")
	}
	if raw, _ := os.ReadFile("README.md"); strings.ReplaceAll(string(raw), "\r\n", "\n") != "hello\n" {
		t.Fatalf("README: %q", raw)
	}
	got := commits(t, g)
	if len(got) != 3 || got[0] != "chore: undo add lib billing" {
		t.Fatalf("commits: %q", got)
	}
	if tree, _ := g.Run("diff", "--stat", before, "HEAD"); tree != "" {
		t.Fatalf("the tree differs from the original:\n%s", tree)
	}
	if _, err := Prepare(g.Dir); err == nil {
		t.Fatal("nothing is left to undo")
	}
}

func TestUndoRefusesWhenALaterCommitChangedTheFiles(t *testing.T) {
	g := scratch(t, true)
	execute(t, "billing")
	put(t, g.Dir, "README.md", "hello\nbilling\nmore\n")
	if err := g.CommitAll("docs: edit"); err != nil {
		t.Fatal(err)
	}
	_, err := Prepare(g.Dir)
	if err == nil || !strings.Contains(err.Error(), "files changed since") {
		t.Fatalf("got %v", err)
	}
}

func TestUndoRefusesUncommittedChanges(t *testing.T) {
	g := scratch(t, true)
	execute(t, "billing")
	put(t, g.Dir, "libs/billing/index.txt", "edited\n")
	_, err := Prepare(g.Dir)
	if err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("got %v", err)
	}
}

func TestUndoWithoutAnOperationSaysHowToStart(t *testing.T) {
	g := scratch(t, true)
	_, err := Prepare(g.Dir)
	if err == nil || !strings.Contains(err.Error(), "commit_operations") {
		t.Fatalf("got %v", err)
	}
}

func TestUndoStepsBackThroughOperations(t *testing.T) {
	g := scratch(t, true)
	execute(t, "one")
	// A second operation edits README.md again, so only the last can be undone.
	execute(t, "two")
	c, err := Prepare(g.Dir)
	if err != nil || c.Record.Command != "add lib two" {
		t.Fatalf("got %+v %v", c.Record, err)
	}
	if _, err := Undo(g.Dir, c); err != nil {
		t.Fatal(err)
	}
	c, err = Prepare(g.Dir)
	if err != nil || c.Record.Command != "add lib one" {
		t.Fatalf("got %+v %v", c.Record, err)
	}
}
