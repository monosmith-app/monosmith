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
