package pipeline

import (
	"os"
	"strings"
	"testing"
)

func TestAnotherGolangciLintHasASmithFixThatRunsCISetup(t *testing.T) {
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
		if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 1 || strings.Join(f.Fix.Steps[0].Argv, " ") != "smith ci setup" {
			t.Fatalf("the remedy names smith ci setup, so the fix runs it: %+v", f.Fix)
		}
		return
	}
	t.Fatal("no golangci-lint finding")
}
