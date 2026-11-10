package cipipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/release"
)

func TestAReleaseThatReleasedSomethingTellsTheWorkflowEvenWhenItFailed(t *testing.T) {
	out := filepath.Join(t.TempDir(), "output")
	t.Setenv("GITHUB_OUTPUT", out)
	markReleased(release.Report{Results: []release.ReportResult{{Status: "failed"}}})
	if _, err := os.Stat(out); err == nil {
		t.Fatal("nothing was released: no output")
	}
	partial := release.Report{Error: "the seventh failed", Results: []release.ReportResult{{Status: "released"}, {Status: "released"}, {Status: "failed"}}}
	markReleased(partial)
	got, _ := os.ReadFile(out)
	if string(got) != "released=true\n" {
		t.Fatalf("one line, once: %q", got)
	}
}

func TestMarkReleasedNeedsNoWorkflow(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "")
	markReleased(release.Report{Results: []release.ReportResult{{Status: "released"}}}) // no file, no panic
}
