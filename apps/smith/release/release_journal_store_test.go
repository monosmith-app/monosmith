package release

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTheJournalOnlyGrowsAndIsReadBack(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 10, 19, 12, 0, 0, 0, time.UTC)
	first := ReportResult{Item: Item{Project: "cli", Next: "1.2.0", Tag: "cli@1.2.0"}, Status: "failed", Packages: []string{"acme@1.2.0"}, Error: "RubyGems: boom"}
	second := ReportResult{Item: Item{Project: "cli", Next: "1.2.0", Tag: "cli@1.2.0"}, Status: "released", Pushed: true, ExistingPackages: []string{"acme@1.2.0"}, Packages: []string{"acme-1.2.0.gem"}}
	for _, res := range []ReportResult{first, second} {
		if err := AppendJournal(root, "cli", JournalEntries(res, at)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "dist", "release", "cli", "journal.json")); err != nil {
		t.Fatal(err)
	}
	j, err := ReadJournal(root, "cli")
	if err != nil {
		t.Fatal(err)
	}
	var steps []string
	for _, e := range j.Entries {
		steps = append(steps, e.Step+":"+e.Outcome)
	}
	got := strings.Join(steps, " ")
	for _, expect := range []string{"package:published", "release:failed", "package:existing", "tag:done", "push:done"} {
		if !strings.Contains(got, expect) {
			t.Errorf("journal %q lacks %s", got, expect)
		}
	}
	if len(j.Entries) != 6 {
		t.Errorf("both runs are kept, got %d entries: %s", len(j.Entries), got)
	}
	if !JournalHas(j, "acme", "1.2.0") || JournalHas(j, "acme", "9.9.9") || JournalHas(j, "other", "1.2.0") {
		t.Error("JournalHas should match a published package name and version only")
	}
}

func TestAProjectThatNeverReleasedHasAnEmptyJournal(t *testing.T) {
	root := t.TempDir()
	j, err := ReadJournal(root, "cli")
	if err != nil || len(j.Entries) != 0 {
		t.Fatalf("%+v %v", j, err)
	}
	journals, err := ReadJournals(root)
	if err != nil || len(journals) != 0 {
		t.Fatalf("%+v %v", journals, err)
	}
	if got := RenderJournalStatus(nil); !strings.Contains(got, "No release journal") {
		t.Errorf("status: %q", got)
	}
}

func TestStatusListsEachProjectsStepsWithTheFailure(t *testing.T) {
	root := t.TempDir()
	at := time.Date(2026, 10, 19, 12, 30, 0, 0, time.UTC)
	res := ReportResult{Item: Item{Project: "cli", Next: "1.2.0", Tag: "cli@1.2.0"}, Status: "failed", Packages: []string{"acme@1.2.0"}, Error: "RubyGems: boom"}
	if err := AppendJournal(root, "cli", JournalEntries(res, at)); err != nil {
		t.Fatal(err)
	}
	journals, err := ReadJournals(root)
	if err != nil {
		t.Fatal(err)
	}
	got := RenderJournalStatus(journals)
	for _, expect := range []string{"cli\n", "2026-10-19 12:30", "package", "published", "acme@1.2.0", "failed", "(RubyGems: boom)"} {
		if !strings.Contains(got, expect) {
			t.Errorf("status lacks %q:\n%s", expect, got)
		}
	}
}
