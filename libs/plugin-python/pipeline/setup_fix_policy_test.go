package pipeline

import (
	"strings"
	"testing"
)

func TestTheVenvFixCreatesTheVenvAndInstallsTheToolchainWithoutSmithCISetup(t *testing.T) {
	if setupFix(false, "ws") != nil {
		t.Fatal("an existing venv needs no fix")
	}
	fix := setupFix(true, "ws")
	if len(fix.Steps) != 2 {
		t.Fatalf("two steps expected: %+v", fix.Steps)
	}
	for _, s := range fix.Steps {
		if s.Argv[0] == "smith" {
			t.Fatalf("the fix must not run smith: %v", s.Argv)
		}
	}
	if got := strings.Join(fix.Steps[0].Argv, " "); !strings.Contains(got, "-m venv") || !strings.HasSuffix(got, ".venv") {
		t.Fatalf("create step: %q", got)
	}
	if got := strings.Join(fix.Steps[1].Argv, " "); !strings.Contains(got, "pip install") || !strings.HasSuffix(got, "requirements-dev.txt") {
		t.Fatalf("install step: %q", got)
	}
}
