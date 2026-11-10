package cipipeline

import (
	"strings"
	"testing"
)

func TestAuditConclusionCountsSkippedChecks(t *testing.T) {
	got := auditConclusion([]stepOutcome{{Name: "npm audit"}, {Name: "dotnet audit", Verdict: stepSkipped}}, 2)
	if !strings.Contains(got, "1 skipped because a tool is missing (not audited)") {
		t.Errorf("conclusion: %q", got)
	}
}
