package convert

import (
	"strings"
	"testing"
)

func TestReportKeepsChangesToFilesApartFromNotes(t *testing.T) {
	plan := manyProjectsPlan(1)
	plan.AddEdit("", "turbo.json", "removed: its tasks are the package.json scripts")
	out := Report(plan, ReportOptions{DryRun: true})
	edits := strings.Index(out, "Changes to your files (1):")
	notes := strings.Index(out, "Converted, nothing to do (")
	if edits < 0 || notes < 0 || edits > notes {
		t.Fatalf("the sections are missing or in the wrong order:\n%s", out)
	}
	if at := strings.Index(out, "turbo.json: removed"); at < edits || at > notes {
		t.Fatalf("turbo.json is not under the changes:\n%s", out)
	}
}
