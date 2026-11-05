package pipeline

import (
	"os"
	"strings"
	"testing"
)

// The fix of a missing golangci-lint is a question; whatever answer runs a
// command runs the narrow install of the pinned version, never smith ci setup.
func TestAnotherGolangciLintHasAFixThatRunsOnlyTheInstallOfThePin(t *testing.T) {
	realProbe, realLook := probe, lookAll
	defer func() { probe, lookAll = realProbe, realLook }()
	lookAll = func(string) []string { return nil }
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "go" {
			return "go1.26.0", nil
		}
		return "", os.ErrNotExist
	}
	for _, f := range CheckToolVersions(t.TempDir(), "1.24", true) {
		if !strings.HasPrefix(f.Check, "golangci-lint") {
			continue
		}
		if f.Fix == nil || f.Fix.Safe || len(f.Fix.Choices) != 1 {
			t.Fatalf("the fix is a question, never applied unasked: %+v", f.Fix)
		}
		ran := 0
		for _, option := range f.Fix.Choices[0].Options {
			for _, step := range option.Fix.Steps {
				ran++
				if strings.Join(step.Argv, " ") != "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+GolangciLintVersion {
					t.Fatalf("the answer runs the install of the pinned version, not smith ci setup: %+v", step)
				}
			}
		}
		if ran != 2 {
			t.Fatalf("two ways to install it: %d", ran)
		}
		return
	}
	t.Fatal("no golangci-lint finding")
}
