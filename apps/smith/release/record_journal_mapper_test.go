package release

import (
	"testing"
	"time"
)

func TestASkippedPrereleaseIsJournaledAsSkipped(t *testing.T) {
	entries := JournalEntries(ReportResult{SkippedPackages: []string{"PyPI acme: skipped: pre-release"}}, time.Time{})
	if len(entries) != 1 || entries[0].Step != "package" || entries[0].Outcome != OutcomeSkipped {
		t.Fatalf("%+v", entries)
	}
}
