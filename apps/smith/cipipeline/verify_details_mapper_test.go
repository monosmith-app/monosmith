package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func detailProject(name, root string) *projectgraph.Project {
	return &projectgraph.Project{Project: protocol.Project{Name: name, Root: root}}
}

func TestVerifyDetailsShowTestsLintBuildAndCoverageOfEachProject(t *testing.T) {
	ws := t.TempDir()
	writeFile(t, filepath.Join(ws, "libs", "core", "coverage", "coverage-summary.json"), `{"total":{"lines":{"pct":81.25}}}`)
	writeFile(t, filepath.Join(ws, "libs", "core", "test-results", "junit.xml"), "<testsuites/>")
	writeFile(t, filepath.Join(ws, "dist", "libs", "core", "index.js"), strings.Repeat("x", 4096))
	outputs := newTaskOutputs()
	for _, line := range []string{
		"core:test │ Tests:       42 passed, 42 total",
		"core:lint │ ✖ 0 problems",
		"api:test │ Tests:       3 failed, 9 passed, 12 total",
		"api:lint │ ✖ 7 problems (7 errors, 0 warnings)",
	} {
		_, _ = outputs.Write([]byte(line + "\n"))
	}
	results := []taskrun.Result{
		{Task: "core:lint", Status: taskrun.Succeeded}, {Task: "core:test", Status: taskrun.Succeeded}, {Task: "core:build", Status: taskrun.Succeeded},
		{Task: "api:lint", Status: taskrun.Failed}, {Task: "api:test", Status: taskrun.Failed},
	}
	projects := []*projectgraph.Project{detailProject("core", "libs/core"), detailProject("api", "apps/api")}
	icons := theme.Unicode

	d := reportDetails(results, outputs, projects, ws, time.Now().Add(-time.Hour), icons)
	text := strings.Join(d.Facts.Text("more"), "\n")
	for _, want := range []string{"Project", "Tests", "Lint", "Build", "Coverage", "× 3 failed, 9 passed", "× 7 issues", "✓ 42 tests passed", "✓ passed", "4.0 KB", "81.2%"} {
		if !strings.Contains(text, want) {
			t.Errorf("the facts table lacks %q:\n%s", want, text)
		}
	}
	if rows := d.Facts.Rows; len(rows) != 2 || rows[0].Cells[0] != "api" {
		t.Errorf("a failing project must come first: %+v", rows)
	}
	paths := strings.Join(d.Paths.Text("more"), "\n")
	for _, want := range []string{"dist/libs/core", "libs/core/test-results/junit.xml", "libs/core/coverage/coverage-summary.json"} {
		if !strings.Contains(paths, want) {
			t.Errorf("the paths table lacks %q:\n%s", want, paths)
		}
	}
	if strings.Contains(paths, "api") {
		t.Errorf("a project with no path to show has a row:\n%s", paths)
	}
	if d.CoverageNote != "" {
		t.Errorf("a coverage report is in place, yet: %s", d.CoverageNote)
	}
}

func TestVerifyDetailsNeverTurnCoverageOnAndSayWhenThereIsNone(t *testing.T) {
	ws := t.TempDir()
	outputs := newTaskOutputs()
	results := []taskrun.Result{{Task: "web:test", Status: taskrun.Cached}}
	d := reportDetails(results, outputs, []*projectgraph.Project{detailProject("web", "apps/web")}, ws, time.Now(), theme.Unicode)
	text := strings.Join(d.Facts.Text("more"), "\n")
	if !strings.Contains(text, "≈") || !strings.Contains(text, "none") {
		t.Errorf("a cached test and a missing lint are not shown:\n%s", text)
	}
	if !strings.Contains(d.CoverageNote, "not collected") {
		t.Errorf("no note that coverage is not collected: %q", d.CoverageNote)
	}
	if len(d.Paths.Rows) != 0 {
		t.Errorf("paths with nothing in place: %+v", d.Paths.Rows)
	}
}

func TestACoverageReportOlderThanTheRunShowsItsAge(t *testing.T) {
	ws := t.TempDir()
	path := filepath.Join(ws, "apps", "web", "coverage", "coverage-summary.json")
	writeFile(t, path, `{"total":{"lines":{"pct":50}}}`)
	old := time.Now().Add(-49 * time.Hour)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	results := []taskrun.Result{{Task: "web:test", Status: taskrun.Succeeded}}
	d := reportDetails(results, newTaskOutputs(), []*projectgraph.Project{detailProject("web", "apps/web")}, ws, time.Now(), theme.Unicode)
	if text := strings.Join(d.Facts.Text("more"), "\n"); !strings.Contains(text, "50.0% (2d old)") {
		t.Errorf("an old report is passed off as new:\n%s", text)
	}
}
