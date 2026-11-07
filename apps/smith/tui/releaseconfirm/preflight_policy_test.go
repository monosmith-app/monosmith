package releaseconfirm

import (
	"strings"
	"testing"
)

func ready() Facts {
	return Facts{DoctorKnown: true, Tests: TestsPassed, GitKnown: true, Branch: "main", Remote: "origin", HasRemote: true, Upstream: true}
}

func TestAReadyWorkspaceHasNoWarning(t *testing.T) {
	checks := Preflight(ready())
	if Warns(checks) {
		t.Fatalf("warns: %+v", checks)
	}
	want := []string{"Doctor: no findings", "Tests: the last test run passed", "Branch main: pushed"}
	for i, c := range checks {
		if c.Text != want[i] {
			t.Fatalf("check %d is %q, want %q", i, c.Text, want[i])
		}
	}
}

func TestEachFailingItemIsAWarningThatSaysWhy(t *testing.T) {
	cases := []struct {
		name   string
		change func(*Facts)
		index  int
		want   string
	}{
		{"findings", func(f *Facts) { f.DoctorFindings = 2 }, 0, "Doctor: 2 findings are open"},
		{"one finding", func(f *Facts) { f.DoctorFindings = 1 }, 0, "Doctor: 1 finding is open"},
		{"doctor not read", func(f *Facts) { f.DoctorKnown = false }, 0, "Doctor: not checked yet"},
		{"tests failed", func(f *Facts) { f.Tests = TestsFailed }, 1, "Tests: the last test run failed"},
		{"tests not run", func(f *Facts) { f.Tests = TestsNone }, 1, "Tests: no test run in this session"},
		{"ahead", func(f *Facts) { f.Ahead = 3 }, 2, "Branch main: 3 commits are not pushed"},
		{"ahead by one", func(f *Facts) { f.Ahead = 1 }, 2, "Branch main: 1 commit is not pushed"},
		{"no upstream", func(f *Facts) { f.Upstream = false }, 2, "Branch main: not pushed yet, it has no upstream"},
		{"no remote", func(f *Facts) { f.HasRemote = false }, 2, "Branch main: no remote, so it is not pushed"},
		{"git not read", func(f *Facts) { f.GitKnown = false }, 2, "Branch: not known"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := ready()
			c.change(&f)
			got := Preflight(f)[c.index]
			if !got.Warn || !strings.Contains(got.Text, c.want) {
				t.Fatalf("got %+v, want a warning %q", got, c.want)
			}
		})
	}
}

func TestTheRunLabelSaysAnywayOnlyWhenACheckWarns(t *testing.T) {
	if got := RunLabel(Preflight(ready())); got != "Release now" {
		t.Fatalf("clean: %q", got)
	}
	f := ready()
	f.Tests = TestsFailed
	if got := RunLabel(Preflight(f)); got != "Release anyway" {
		t.Fatalf("warning: %q", got)
	}
}
