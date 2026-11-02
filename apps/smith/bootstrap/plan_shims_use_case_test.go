package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

func TestPlanShimsCreatesThenDiffsThenIsEmpty(t *testing.T) {
	root := t.TempDir()
	plan, err := PlanShims(root)
	if err != nil || plan.Empty() {
		t.Fatalf("%v %v", plan, err)
	}
	var b strings.Builder
	operationplan.Render(&b, plan)
	if !strings.Contains(b.String(), "create  "+ShellPath) {
		t.Fatalf("a missing launcher is created:\n%s", b.String())
	}
	if _, statErr := os.Stat(filepath.Join(root, ShellPath)); statErr == nil {
		t.Fatal("PlanShims must not write")
	}
	if err := os.MkdirAll(filepath.Join(root, filepath.Dir(ShellPath)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ShellPath), []byte("#!/bin/sh\necho my own launcher\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	plan, _ = PlanShims(root)
	b.Reset()
	operationplan.Render(&b, plan)
	if !strings.Contains(b.String(), "change  "+ShellPath) || !strings.Contains(b.String(), "-echo my own launcher") {
		t.Fatalf("a launcher that differs shows what is lost:\n%s", b.String())
	}
	if _, err := WriteShims(root, false); err != nil {
		t.Fatal(err)
	}
	if plan, _ = PlanShims(root); !plan.Empty() {
		t.Fatal("after writing, the plan is empty")
	}
}
