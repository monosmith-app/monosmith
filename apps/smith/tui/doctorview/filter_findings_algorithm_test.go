package doctorview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// warned is the sample report with one warning added to web's group.
func warned() Model {
	m := sample()
	r := *m.Report
	r.Findings = append(append([]doctor.Finding(nil), r.Findings...), doctor.Finding{Finding: protocol.Finding{Check: "Docker of web is old", Warning: true, Project: "web", Detail: "old", Remedy: "update"}, Source: "docker"})
	r.Warnings = 1
	m.Report = &r
	return m
}

func TestAFilterHidesItsStateInEveryGroupAndHeadingsKeepTheirCounts(t *testing.T) {
	m := warned()
	if out := plain(m); !strings.Contains(out, "Docker registries of web agree") || !strings.Contains(out, "Docker of web is old") {
		t.Fatalf("every filter starts on:\n%s", out)
	}
	m = m.Update(key("x"))
	out := plain(m)
	if strings.Contains(out, "Docker registries of web agree") || strings.Contains(out, "Go is installed") || !strings.Contains(out, "Docker of web is old") {
		t.Fatalf("x hides the failed checks of every group and keeps the warning:\n%s", out)
	}
	if !strings.Contains(out, "web  1 failed · 1 warning") {
		t.Fatalf("the heading counts the whole group, not only the rows drawn:\n%s", out)
	}
	m = m.Update(key("x")).Update(key("w"))
	if out = plain(m); strings.Contains(out, "Docker of web is old") || !strings.Contains(out, "Docker registries of web agree") {
		t.Fatalf("w hides the warnings, x brought the failures back:\n%s", out)
	}
	if m.Shows(FilterWarnings) || !m.Shows(FilterFailed) || m.Shows(FilterPassed) {
		t.Fatalf("Shows says what is listed: %+v", m)
	}
}

func TestThePassedFilterListsThePassedChecksAndTheDefaultIsTheCollapsedLine(t *testing.T) {
	m := warned()
	if out := plain(m); strings.Contains(out, "git is installed") || !strings.Contains(out, "1 checks passed") {
		t.Fatalf("the passed checks start behind one line:\n%s", out)
	}
	if out := plain(m.Update(key("p"))); !strings.Contains(out, "git is installed") || strings.Contains(out, "1 checks passed") {
		t.Fatalf("p lists the passed checks in place of the line:\n%s", out)
	}
}

func TestWithEveryFilterOffThePageNamesItsEmptyStateAndTheKeysBringChecksBack(t *testing.T) {
	m := warned().Update(key("x")).Update(key("w"))
	out := plain(m)
	if !strings.Contains(out, "No checks match the filters. Turn one on (x, w or p).") || !m.NothingMatches() {
		t.Fatalf("the empty state is named:\n%s", out)
	}
	if got := m.Update(key("x")); got.NothingMatches() || !strings.Contains(plain(got), "Docker registries of web agree") {
		t.Fatal("with no row to rest on, the key still turns the filter on")
	}
}

func TestAFilterSurvivesTheNextCheckOfTheSameView(t *testing.T) {
	m := warned().Update(key("x"))
	next := m.Replace(Result{Report: *m.Report, Projects: m.Projects})
	if next.Shows(FilterFailed) {
		t.Fatal("a check that finishes while the page is open keeps the person's filters")
	}
}

func TestTheHeadlineHasThreeFilterButtonsWithTheCountsAndTheirKeys(t *testing.T) {
	m := warned()
	lines, _ := m.View(plainTheme, 120, true)
	icons := plainTheme.Icons
	for _, want := range []string{"[" + icons.Failed + " 3 failed (x)]", "[" + icons.Warning + " 1 warning (w)]", "[" + icons.Passed + " 1 passed (p)]"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("headline %q lacks %q", lines[0], want)
		}
	}
	m.Report.Warnings = 2
	if lines, _ = m.View(plainTheme, 120, true); !strings.Contains(lines[0], "2 warnings (w)") {
		t.Errorf("a plural count says warnings: %q", lines[0])
	}
}

func TestEachFilterButtonSpanCoversItsDrawnText(t *testing.T) {
	m := warned()
	for _, w := range []int{120, 30} {
		lines, _ := m.View(plainTheme, w, true)
		spans := m.HeadlineSpans(plainTheme, w)
		seen := 0
		for line, row := range spans {
			cells := []rune(lines[line])
			for _, sp := range row {
				seen++
				if cells[sp.From] != '[' || cells[sp.To-1] != ']' {
					t.Errorf("width %d: span %+v of %q does not frame a button", w, sp, lines[line])
				}
			}
		}
		if seen != 3 {
			t.Errorf("width %d: %d buttons, want 3", w, seen)
		}
	}
	if narrow := m.HeadlineSpans(plainTheme, 30); len(narrow) < 2 {
		t.Errorf("buttons that do not fit move to the next line: %v", narrow)
	}
}
