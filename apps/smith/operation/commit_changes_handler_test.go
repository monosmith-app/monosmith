package operation

import (
	"bytes"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
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

// withOwnWork leaves a change of the person's own in the work tree, then runs an
// operation that does not commit, so smith records which files it changed.
func withOwnWork(t *testing.T) *gitclient.Git {
	t.Helper()
	g := scratch(t, false)
	put(t, ".", "mine.txt", "my own work\n")
	execute(t, "core")
	return g
}

func runCommit(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	cmd := CommitCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestAnOperationRemembersTheFilesItChangedAndNotThePersonsOwn(t *testing.T) {
	g := withOwnWork(t)
	u, ok := ReadPending(g.Dir)
	if !ok || u.Command != "add lib core" || strings.Join(u.Files, ",") != "README.md,libs/core/index.txt" {
		t.Fatalf("recorded %+v %v", u, ok)
	}
}

func TestOnlyOperationCommitsTheOperationsFilesAndLeavesTheOthers(t *testing.T) {
	g := withOwnWork(t)
	if _, err := runCommit(t, "", "--only-operation", "--yes"); err != nil {
		t.Fatal(err)
	}
	files, _ := g.Run("show", "--name-only", "--format=", "HEAD")
	if strings.Contains(files, "mine.txt") || !strings.Contains(files, "libs/core/index.txt") {
		t.Fatalf("the commit holds %q", files)
	}
	if status, _ := g.Run("status", "--porcelain"); !strings.Contains(status, "mine.txt") {
		t.Fatalf("the person's file must stay changed: %q", status)
	}
	if log, _ := ReadLog(g.Dir); log.Last() == nil || log.Last().Command != "add lib core" {
		t.Fatalf("undo must find it: %+v", log)
	}
	if _, ok := ReadPending(g.Dir); ok {
		t.Fatal("a committed operation is no longer pending")
	}
}

func TestAllCommitsEverythingGitShows(t *testing.T) {
	g := withOwnWork(t)
	if _, err := runCommit(t, "", "--all", "--yes"); err != nil {
		t.Fatal(err)
	}
	if status, _ := g.Run("status", "--porcelain"); status != "" {
		t.Fatalf("left over: %q", status)
	}
}

func TestWithoutATerminalAndNoFlagItRefusesAndNamesTheFlags(t *testing.T) {
	g := withOwnWork(t)
	_, err := runCommit(t, "", "--yes")
	if err == nil || !strings.Contains(err.Error(), "--only-operation") || !strings.Contains(err.Error(), "--all") {
		t.Fatalf("error %v", err)
	}
	if subject, _ := g.Run("log", "-1", "--format=%s"); subject != "chore: start" {
		t.Fatalf("nothing may be committed: %q", subject)
	}
}

func TestATerminalAsksAndTheDefaultIsTheOperationsFilesOnly(t *testing.T) {
	g := withOwnWork(t)
	t.Setenv("SMITH_FORCE_INTERACTIVE", "1")
	out, err := runCommit(t, "\ny\n")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{"mine.txt", "1) only the 2 file(s) this operation changed (default)", "2) everything git shows changed (3 files)", "3) cancel"} {
		if !strings.Contains(out, want) {
			t.Errorf("the question lacks %q:\n%s", want, out)
		}
	}
	if files, _ := g.Run("show", "--name-only", "--format=", "HEAD"); strings.Contains(files, "mine.txt") {
		t.Fatalf("the default must leave the person's file out: %q", files)
	}
}

func TestCancellingTheChoiceCommitsNothing(t *testing.T) {
	g := withOwnWork(t)
	t.Setenv("SMITH_FORCE_INTERACTIVE", "1")
	if _, err := runCommit(t, "3\n"); err != nil {
		t.Fatal(err)
	}
	if subject, _ := g.Run("log", "-1", "--format=%s"); subject != "chore: start" {
		t.Fatalf("committed %q", subject)
	}
}

func TestADryRunListsBothSetsAndChangesNothing(t *testing.T) {
	g := withOwnWork(t)
	out, err := runCommit(t, "", "--dry-run")
	if err != nil || !strings.Contains(out, "mine.txt") || !strings.Contains(out, "Dry run") {
		t.Fatalf("%v\n%s", err, out)
	}
	if subject, _ := g.Run("log", "-1", "--format=%s"); subject != "chore: start" {
		t.Fatal("a dry run commits nothing")
	}
}
