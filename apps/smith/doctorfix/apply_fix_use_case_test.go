package doctorfix

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TestHelperProcess is the command the steps of these tests run: the test
// binary itself, told what to do after "--".
func TestHelperProcess(t *testing.T) {
	if os.Getenv("DOCTORFIX_HELPER") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	switch args[0] {
	case "touch":
		_ = os.WriteFile(args[1], []byte("ran"), 0o644)
	case "say":
		fmt.Println("first line")
		fmt.Println("added 12 packages")
	case "fail":
		fmt.Println("the build exploded")
		os.Exit(3)
	}
	os.Exit(0)
}

func helper(t *testing.T, description string, args ...string) protocol.FixStep {
	t.Helper()
	t.Setenv("DOCTORFIX_HELPER", "1")
	return protocol.FixStep{Description: description, Argv: append([]string{os.Args[0], "-test.run=TestHelperProcess", "--"}, args...)}
}

func TestApplyWritesTheFilesThenRunsTheSteps(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "go.work"), []byte("go 1.26\n"), 0o644)
	fix := protocol.Fix{Description: "fix go.work", Safe: true,
		Files: []protocol.FileEdit{{Path: "go.work", Content: "go 1.26\n\nuse ./api\n"}},
		Steps: []protocol.FixStep{helper(t, "leave a mark", "touch", filepath.Join(root, "mark"))},
	}
	var shown bytes.Buffer
	if err := ShowFix(&shown, root, Planned{Fix: fix}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"+use ./api", "step 1: leave a mark"} {
		if !strings.Contains(shown.String(), want) {
			t.Fatalf("the fix shows %q:\n%s", want, shown.String())
		}
	}
	if err := Apply(context.Background(), root, fix, io.Discard); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "go.work")); string(got) != "go 1.26\n\nuse ./api\n" {
		t.Fatalf("go.work %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, "mark")); err != nil {
		t.Fatalf("the step ran: %v", err)
	}
}

func TestApplyStopsAtTheFirstFailingStepAndSaysWhich(t *testing.T) {
	root := t.TempDir()
	fix := protocol.Fix{Description: "three steps", Steps: []protocol.FixStep{
		helper(t, "first", "touch", filepath.Join(root, "one")),
		helper(t, "second", "fail"),
		helper(t, "third", "touch", filepath.Join(root, "three")),
	}}
	err := Apply(context.Background(), root, fix, io.Discard)
	var step *StepError
	if !errors.As(err, &step) || step.Number != 2 || step.Total != 3 || !strings.Contains(step.Output, "the build exploded") || !strings.Contains(err.Error(), "step 2 of 3") {
		t.Fatalf("the second step failed with its output: %v %+v", err, step)
	}
	if _, err := os.Stat(filepath.Join(root, "one")); err != nil {
		t.Fatal("the first step ran")
	}
	if _, err := os.Stat(filepath.Join(root, "three")); err == nil {
		t.Fatal("the third step must not run after a failure")
	}
}

func TestAnAdvisoryStepThatFailsIsShownAndTheNextStepsStillRun(t *testing.T) {
	root := t.TempDir()
	lint := helper(t, "lint", "fail")
	lint.Advisory = true
	fix := protocol.Fix{Description: "check", Steps: []protocol.FixStep{lint, helper(t, "doctor", "touch", filepath.Join(root, "doctor"))}}
	var out bytes.Buffer
	if err := Apply(context.Background(), root, fix, &out); err != nil {
		t.Fatalf("an advisory failure stops nothing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "doctor")); err != nil {
		t.Fatal("the step after the advisory one ran")
	}
	for _, want := range []string{"▲ lint\n", "the build exploded\n", "✓ doctor\n"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "✓ lint") {
		t.Fatalf("a failed advisory step is not marked passed:\n%s", out.String())
	}
	findings, err := ApplyWithFindings(context.Background(), root, fix, io.Discard)
	if err != nil || len(findings) != 1 || findings[0].Step.Description != lint.Description || !strings.Contains(findings[0].Output, "the build exploded") {
		t.Fatalf("the failed advisory step is a finding with its output: %+v, %v", findings, err)
	}
}

func TestApplyRefusesWhatLeavesTheWorkspaceOrStillAsks(t *testing.T) {
	root := t.TempDir()
	if err := Apply(context.Background(), root, protocol.Fix{Files: []protocol.FileEdit{{Path: "../outside.txt", Content: "x"}}}, io.Discard); err == nil {
		t.Fatal("a fix must stay in the workspace")
	}
	if err := Apply(context.Background(), root, protocol.Fix{Steps: []protocol.FixStep{{Argv: []string{"x"}, Dir: ".."}}}, io.Discard); err == nil {
		t.Fatal("a step must run inside the workspace")
	}
	asks := protocol.Fix{Description: "asks", Choices: []protocol.FixChoice{{Question: "which?"}}}
	if err := Apply(context.Background(), root, asks, io.Discard); err == nil {
		t.Fatal("a fix with questions must be resolved before it applies")
	}
}

func TestApplyNamesAMissingProgram(t *testing.T) {
	err := Apply(context.Background(), t.TempDir(), protocol.Fix{Steps: []protocol.FixStep{{Argv: []string{"no-such-program-doctorfix"}}}}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "no-such-program-doctorfix") {
		t.Fatalf("%v", err)
	}
}

// notes collects the detail lines Apply hands to an output that wants them.
type notes struct {
	bytes.Buffer
	lines []string
}

func (n *notes) Note(line string) { n.lines = append(n.lines, line) }

func TestApplyTellsANoterWhatItWritesAndWhatTheCommandPrints(t *testing.T) {
	root := t.TempDir()
	fix := protocol.Fix{Description: "set the dependency",
		Files: []protocol.FileEdit{{Path: "package.json", Content: "{}"}},
		Steps: []protocol.FixStep{helper(t, "install", "say")},
	}
	var out notes
	if err := Apply(context.Background(), root, fix, &out); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(out.lines, "|"); got != "Writing package.json|first line|added 12 packages" {
		t.Fatalf("notes: %q", got)
	}
	if strings.Contains(out.String(), "added 12 packages") {
		t.Fatalf("command output must not flood the plain output: %q", out.String())
	}
}

func TestApplyFilesRemovesAFolderTheDeletedFileLeavesEmpty(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"a/scripts/one.mjs", "b/scripts/one.mjs", "b/scripts/keep.mjs"} {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fix := protocol.Fix{Files: []protocol.FileEdit{{Path: "a/scripts/one.mjs", Delete: true}, {Path: "b/scripts/one.mjs", Delete: true}}}
	if err := ApplyFiles(root, fix); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "a", "scripts")); err == nil {
		t.Error("an empty folder is removed with its last file")
	}
	if _, err := os.Stat(filepath.Join(root, "b", "scripts", "keep.mjs")); err != nil {
		t.Error("a folder that still holds a file stays")
	}
}
