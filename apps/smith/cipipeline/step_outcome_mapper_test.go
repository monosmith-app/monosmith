package cipipeline

import (
	"strings"
	"testing"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

func TestTheStepTableSaysWhichChecksRanWereUnchangedOrReportedProblems(t *testing.T) {
	outcomes := []stepOutcome{
		{Name: "govulncheck smith", Took: 12 * time.Second},
		{Name: "pip-audit", Verdict: stepUnchanged, Since: "2026-10-26"},
		{Name: "bundler-audit", Verdict: stepProblem, Took: 2 * time.Second},
	}
	got := strings.Join(stepTable("Check", outcomes, theme.Unicode).Text("checks done"), "\n")
	for _, want := range []string{"Check", "govulncheck smith  ✓ done", "↷ unchanged", "nothing changed since it passed on 2026-10-26", "▲ problems", "reported problems, not blocking"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if conclusion := auditConclusion(outcomes, 23); conclusion != "3 checks over 23 projects: 1 ran, 1 unchanged since an earlier pass, 1 reported problems (not blocking)" {
		t.Errorf("conclusion: %s", conclusion)
	}
}
