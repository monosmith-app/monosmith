package doctorfix

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func command(p Planned, answers []int) string {
	return fmt.Sprintf("smith doctor --fix --only %q --choose %v", p.Description, answers)
}

func planFor(t *testing.T, root string) []Planned {
	t.Helper()
	safe := protocol.Fix{Description: "restore the owned file", Safe: true, Files: []protocol.FileEdit{{Path: "owned.txt", Content: "owned\n"}}}
	asks := protocol.Fix{Description: "oxfmt is missing", Choices: []protocol.FixChoice{{
		Question: "install oxfmt or remove the script?",
		Options: []protocol.FixChoiceOption{
			{Label: "install it", Fix: protocol.Fix{Steps: []protocol.FixStep{helper(t, "install", "touch", filepath.Join(root, "installed"))}}},
			{Label: "remove the script", Fix: protocol.Fix{Files: []protocol.FileEdit{{Path: "script.txt", Content: "removed\n"}}}},
		},
	}}}
	return Group([]Item{{Check: "owned", Fix: &safe}, {Check: "scripts", Fix: &asks}})
}

func TestWalkWithoutATerminalAppliesOnlySafeFixesAndNamesTheCommandForTheRest(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	res, err := Walk(context.Background(), planFor(t, root), Options{Root: root, Out: &out, Yes: true, Command: command})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Applied) != 1 || res.Applied[0] != "restore the owned file" {
		t.Fatalf("only the safe fix: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, "installed")); err == nil {
		t.Fatal("a fix that asks must not run unasked")
	}
	if len(res.Pending) != 1 || len(res.Pending[0].Commands) != 2 || !strings.Contains(res.Pending[0].Commands[1], "[2]") || !strings.Contains(res.Pending[0].Commands[0], "# install it") {
		t.Fatalf("one command for each option: %+v", res.Pending)
	}
	if !strings.Contains(out.String(), "1) install it") {
		t.Fatalf("the options are shown:\n%s", out.String())
	}
}

func TestWalkAppliesANamedFixOnceItsQuestionsAreAnswered(t *testing.T) {
	root := t.TempDir()
	plan := planFor(t, root)[1:]
	res, err := Walk(context.Background(), plan, Options{Root: root, Out: &bytes.Buffer{}, Yes: true, Explicit: true, Choose: []int{2}, Command: command})
	if err != nil || len(res.Applied) != 1 || len(res.Pending) != 0 {
		t.Fatalf("%+v %v", res, err)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "script.txt")); string(got) != "removed\n" {
		t.Fatalf("option 2's file: %q", got)
	}
	res, _ = Walk(context.Background(), plan, Options{Root: root, Out: &bytes.Buffer{}, Yes: true, Choose: []int{1}, Command: command})
	if len(res.Applied) != 0 || len(res.Pending) != 1 || len(res.Pending[0].Commands) != 1 {
		t.Fatalf("without --only the named command is still printed, not run: %+v", res)
	}
}

func TestWalkInATerminalAsksNumberedQuestionsThenConfirms(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	// Confirm the safe batch, pick option 1, confirm the fix.
	res, err := Walk(context.Background(), planFor(t, root), Options{Root: root, Out: &out, In: strings.NewReader("y\n1\ny\n"), Interactive: true})
	if err != nil || len(res.Applied) != 2 {
		t.Fatalf("%+v %v\n%s", res, err, out.String())
	}
	if _, err := os.Stat(filepath.Join(root, "installed")); err != nil {
		t.Fatalf("option 1 ran its step: %v", err)
	}
	if !strings.Contains(out.String(), "install oxfmt or remove the script?") || !strings.Contains(out.String(), "0) skip") {
		t.Fatalf("the numbered prompt:\n%s", out.String())
	}
}

func TestWalkInATerminalLetsTheUserSkipOrDecline(t *testing.T) {
	root := t.TempDir()
	res, err := Walk(context.Background(), planFor(t, root)[1:], Options{Root: root, Out: &bytes.Buffer{}, In: strings.NewReader("0\n"), Interactive: true})
	if err != nil || len(res.Applied) != 0 || len(res.Skipped) != 1 {
		t.Fatalf("0 skips: %+v %v", res, err)
	}
	res, _ = Walk(context.Background(), planFor(t, root)[1:], Options{Root: root, Out: &bytes.Buffer{}, In: strings.NewReader("2\nn\n"), Interactive: true})
	if len(res.Applied) != 0 || len(res.Skipped) != 1 {
		t.Fatalf("declining the confirmation applies nothing: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(root, "script.txt")); err == nil {
		t.Fatal("nothing was written")
	}
}

func TestWalkStopsAtAFailingStepAndNamesIt(t *testing.T) {
	root := t.TempDir()
	fix := protocol.Fix{Description: "install", Safe: true, Steps: []protocol.FixStep{helper(t, "the install", "fail")}}
	var out bytes.Buffer
	_, err := Walk(context.Background(), Group([]Item{{Check: "c", Fix: &fix}}), Options{Root: root, Out: &out, Yes: true})
	if err == nil || !strings.Contains(err.Error(), "step 1 of 1") || !strings.Contains(out.String(), "the build exploded") {
		t.Fatalf("%v\n%s", err, out.String())
	}
}
