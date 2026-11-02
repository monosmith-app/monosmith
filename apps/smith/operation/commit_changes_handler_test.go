package operation

import (
	"bytes"
	"strings"
	"testing"
)

func TestCommitCommandCommitsWhatChangedAsOneRecordedOperation(t *testing.T) {
	g := scratch(t, false)
	put(t, ".", "libs/core/index.txt", "x\n")
	put(t, ".", "README.md", "changed\n")
	put(t, ".", ".smith/state.json", "{}\n")
	cmd := CommitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"add", "lib", "core", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "chore: add lib core") {
		t.Fatalf("the commit is named after the operation:\n%s", out.String())
	}
	if subject, _ := g.Run("log", "-1", "--format=%s"); subject != "chore: add lib core" {
		t.Fatalf("subject %q", subject)
	}
	log, err := ReadLog(".")
	if err != nil || log.Last() == nil || log.Last().Command != "add lib core" {
		t.Fatalf("smith undo must be able to find it: %+v %v", log, err)
	}
}

func TestCommitCommandDryRunListsTheFilesAndChangesNothing(t *testing.T) {
	g := scratch(t, false)
	put(t, ".", "libs/core/index.txt", "x\n")
	cmd := CommitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"add", "lib", "core", "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "libs/core/index.txt") {
		t.Fatalf("the plan lists the file:\n%s", out.String())
	}
	if status, _ := g.Run("status", "--porcelain"); status == "" {
		t.Fatal("a dry run must leave the changes uncommitted")
	}
}

func TestCommitCommandSaysWhenThereIsNothingToCommit(t *testing.T) {
	scratch(t, false)
	cmd := CommitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--yes"})
	if err := cmd.Execute(); err != nil || !strings.Contains(out.String(), "Nothing to commit") {
		t.Fatalf("%v %q", err, out.String())
	}
}
