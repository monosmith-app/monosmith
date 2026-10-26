package convert

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertplan"
)

func TestTheResultLeadsWithWhatChangedAndWhatNeedsAPerson(t *testing.T) {
	plan := &convertplan.Plan{Findings: []convertplan.Finding{
		{Severity: convertplan.Changed, Project: "a", Subject: "build", Message: "a build moves from Rollup to esbuild"},
		{Severity: convertplan.Changed, Project: "b", Subject: "build", Message: "b build moves from Rollup to esbuild"},
		{Severity: convertplan.Manual, Subject: "format:check", Message: "the script calls a tool that is not installed"},
	}}
	h := HighlightsOf(plan)
	if len(h.Changed) != 1 || !strings.Contains(h.Changed[0], "(2 projects)") {
		t.Fatalf("two projects, one message: %v", h.Changed)
	}
	if len(h.Checks) != 1 || !strings.Contains(h.Checks[0], "format:check") {
		t.Fatalf("checks: %v", h.Checks)
	}
}
