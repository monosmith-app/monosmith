package doctor

import (
	"strings"
	"testing"
	"time"
)

var today = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

func TestAnUnusedLibThatHasSettledIsSuggestedAKind(t *testing.T) {
	f, ok := UnusedLibHint("core", 0, 3, today.AddDate(0, 0, -30), today)
	if !ok || !f.Warning || f.OK {
		t.Fatalf("a warning, not a failure: %+v", f)
	}
	for _, want := range []string{"30 days"} {
		if !strings.Contains(f.Detail, want) {
			t.Errorf("detail %q lacks %q", f.Detail, want)
		}
	}
	for _, want := range []string{"smith kind core app", "smith kind core package", "--dry-run", "smith remove core"} {
		if !strings.Contains(f.Remedy, want) {
			t.Errorf("remedy %q lacks %q", f.Remedy, want)
		}
	}
}

func TestNothingIsSaidWhenThereIsNothingToJudge(t *testing.T) {
	settled := today.AddDate(0, 0, -30)
	for name, c := range map[string]struct {
		dependents, others int
		added              time.Time
	}{
		"the lib is used":               {1, 3, settled},
		"the lib is the only project":   {0, 0, settled},
		"a fresh workspace":             {0, 3, today.AddDate(0, 0, -2)},
		"just under the settling time":  {0, 3, today.Add(-settledAfter + time.Hour)},
		"git does not know when it was": {0, 3, time.Time{}},
	} {
		if f, ok := UnusedLibHint("core", c.dependents, c.others, c.added, today); ok {
			t.Errorf("%s: said %+v", name, f)
		}
	}
}

func TestAPackageNothingPublishesIsSuggestedALib(t *testing.T) {
	f, ok := UnpublishedPackageHint("sdk", true)
	if !ok || !f.Warning || f.OK {
		t.Fatalf("a warning, not a failure: %+v", f)
	}
	if !strings.Contains(f.Remedy, "smith kind sdk lib") || !strings.Contains(f.Remedy, "--dry-run") {
		t.Errorf("remedy %q", f.Remedy)
	}
	if _, ok := UnpublishedPackageHint("sdk", false); ok {
		t.Error("a package a release publishes has nothing to say")
	}
}
