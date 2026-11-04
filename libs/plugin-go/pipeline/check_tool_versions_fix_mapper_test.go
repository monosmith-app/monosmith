package pipeline

import (
	"os"
	"strings"
	"testing"
)

func TestAnotherGolangciLintHasAFixThatRunsOnlyTheInstallOfThePin(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(_ string, argv []string) (string, error) {
		if argv[0] == "go" {
			return "go1.26.0", nil
		}
		return "", os.ErrNotExist
	}
	for _, f := range CheckToolVersions(".", "1.24", true) {
		if !strings.HasPrefix(f.Check, "golangci-lint") {
			continue
		}
		if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 1 || strings.Join(f.Fix.Steps[0].Argv, " ") != "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@"+GolangciLintVersion || !strings.HasPrefix(f.Fix.Steps[0].Description, "go install ") {
			t.Fatalf("the fix runs the install of the pinned version, not smith ci setup: %+v", f.Fix)
		}
		return
	}
	t.Fatal("no golangci-lint finding")
}
