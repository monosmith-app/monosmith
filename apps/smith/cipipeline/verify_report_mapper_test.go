package cipipeline

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/taskrun"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTestRunnersSummariesAreReadFromTheirOutput(t *testing.T) {
	for name, tc := range map[string]struct {
		lines []string
		want  string
	}{
		"pytest": {[]string{"collected 44 items", "====== 42 passed, 2 skipped in 3.1s ======"}, "42 passed"},
		"vitest": {[]string{" Test Files  3 passed (3)", "      Tests  17 passed (17)"}, "17 passed"},
		"dotnet": {[]string{"Passed!  - Failed:     0, Passed:    9, Skipped:     0, Total:     9"}, "9 passed"},
		"go":     {[]string{"ok  \tgithub.com/x/a\t0.3s", "?   \tgithub.com/x/b\t[no test files]", "ok  \tgithub.com/x/c\t(cached)"}, "2 packages ok"},
		"silent": {[]string{"all good"}, ""},
	} {
		var c testCounter
		for _, l := range tc.lines {
			c.add(l)
		}
		if got := c.summary(); got != tc.want {
			t.Errorf("%s: %q, want %q", name, got, tc.want)
		}
	}
}

func TestVerifyReportListsEachProjectWithItsResultsAndTheFailuresFirst(t *testing.T) {
	outputs := newTaskOutputs()
	_, _ = outputs.Write([]byte("core:test │ ====== 42 passed in 3s ======\natlassian:test │ FAIL TestX\natlassian:test │ expected 1 got 2\n"))
	results := []taskrun.Result{
		{Task: "core:lint", Status: taskrun.Succeeded, Duration: 2 * time.Second},
		{Task: "core:test", Status: taskrun.Succeeded, Duration: 10 * time.Second},
		{Task: "core:build", Status: taskrun.Cached},
		{Task: "atlassian:lint", Status: taskrun.Succeeded, Duration: time.Second},
		{Task: "atlassian:test", Status: taskrun.Failed, Duration: 4 * time.Second, Reason: "exit 1"},
		{Task: "atlassian:build", Status: taskrun.Skipped, Reason: "because atlassian:test did not succeed"},
	}
	rep := reportTasks(results, VerifyTasks, outputs, theme.Unicode)
	got := strings.Join(rep.Table.Text("passed"), "\n")
	for _, want := range []string{"Project    Lint  Types   Tests", "atlassian  ✓     ○ none  ✗ failed", "core       ✓     ○ none  ✓ 42 passed", "⚡ cached"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Index(got, "atlassian") > strings.Index(got, "core ") {
		t.Errorf("failures come first:\n%s", got)
	}
	if rep.Conclusion != "2 projects checked, 4 of 6 tasks ok (1 from cache), 42 tests passed" {
		t.Errorf("conclusion: %s", rep.Conclusion)
	}
	lines := strings.Join(failureLines(rep.Failed, outputs), "\n")
	if !strings.Contains(lines, "expected 1 got 2") || !strings.HasSuffix(lines, "→ repeat it with `smith run test -p atlassian`") {
		t.Errorf("failure lines:\n%s", lines)
	}
}

func TestSizesReadLikeSpeech(t *testing.T) {
	for n, want := range map[int64]string{812: "812 B", 4300: "4.2 KB", 13000000: "12.4 MB"} {
		if got := formatSize(n); got != want {
			t.Errorf("%d: %q, want %q", n, got, want)
		}
	}
}
