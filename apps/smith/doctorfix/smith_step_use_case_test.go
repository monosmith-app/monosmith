package doctorfix

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
)

// stub replaces the two functions that would start a process, so no test runs one.
func stub(t *testing.T, plan func([]string) (string, error)) *[][]string {
	t.Helper()
	var asked [][]string
	oldPlan, oldSelf := readPlan, selfPath
	readPlan = func(_ context.Context, _ string, argv []string) (string, error) {
		asked = append(asked, argv)
		return plan(argv)
	}
	selfPath = func() (string, error) { return "/bin/this-smith", nil }
	t.Cleanup(func() { readPlan, selfPath = oldPlan, oldSelf })
	return &asked
}

func TestReviewShowsWhatASmithStepWouldDo(t *testing.T) {
	asked := stub(t, func([]string) (string, error) { return "would move defaults 2025 -> 2026\n", nil })
	p := Planned{}
	p.Description = "move the defaults"
	p.Steps = append(p.Steps, smithfix.Step("move", "upgrade", "--defaults"))
	var out bytes.Buffer
	if err := ShowFix(&out, t.TempDir(), p); err != nil {
		t.Fatal(err)
	}
	want := []string{"/bin/this-smith", "upgrade", "--defaults", "--dry-run"}
	if len(*asked) != 1 || !slices.Equal((*asked)[0], want) {
		t.Fatalf("asked %v, want %v", *asked, want)
	}
	for _, s := range []string{"$ smith upgrade --defaults --yes", "what it would do ($ smith upgrade --defaults --dry-run)", "would move defaults 2025 -> 2026"} {
		if !strings.Contains(out.String(), s) {
			t.Errorf("review lacks %q:\n%s", s, out.String())
		}
	}
}

func TestReviewSaysWhenThePlanCannotBePrinted(t *testing.T) {
	stub(t, func([]string) (string, error) { return "", errors.New("exited with code 1") })
	p := Planned{}
	p.Steps = append(p.Steps, smithfix.Step("regenerate", "ci", "generate"))
	var out bytes.Buffer
	if err := ShowFix(&out, t.TempDir(), p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "smith could not print the plan") {
		t.Fatalf("a plan that failed must be said:\n%s", out.String())
	}
}

func TestOtherProgramsGetNoPlanAndKeepTheirName(t *testing.T) {
	asked := stub(t, func([]string) (string, error) { return "", nil })
	p := Planned{}
	p.Steps = append(p.Steps, smithfix.Step("agent", "agent", "install"))
	p.Steps[0].Argv = []string{"npm", "install"}
	var out bytes.Buffer
	if err := ShowFix(&out, t.TempDir(), p); err != nil {
		t.Fatal(err)
	}
	if len(*asked) != 0 {
		t.Fatalf("asked %v for a step that is not smith", *asked)
	}
	if got := resolveProgram([]string{"npm", "install"}); got[0] != "npm" {
		t.Fatalf("npm became %v", got)
	}
	if got := resolveProgram([]string{"smith", "derive"}); got[0] != "/bin/this-smith" {
		t.Fatalf("smith became %v", got)
	}
}
