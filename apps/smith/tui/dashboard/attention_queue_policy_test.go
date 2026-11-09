package dashboard

import (
	"strings"
	"testing"
	"time"
)

// buttonsOf is "label:command:key" for each button of an item, joined by "|".
func buttonsOf(it Item) string {
	var out []string
	for _, b := range it.Buttons {
		out = append(out, b.Label+":"+b.CommandID+":"+b.Key)
	}
	return strings.Join(out, "|")
}

func kindsOf(q []Item) string {
	var out []string
	for _, it := range q {
		out = append(out, string(it.Kind))
	}
	return strings.Join(out, ",")
}

func TestTheQueueListsOnlyWhatAppliesInTheOrderOfUrgency(t *testing.T) {
	ready := &Ready{Packages: 2, Name: "a", To: "1.0.0"}
	cases := []struct {
		name  string
		f     Facts
		kinds string
	}{
		{"no projects", Facts{ProjectsLoaded: true, Projects: 0, Checked: true}, "empty"},
		{"no projects wins over a doctor that still runs", Facts{ProjectsLoaded: true}, "empty"},
		{"doctor still runs", Facts{ProjectsLoaded: true, Projects: 3, Affected: 2}, "checking"},
		{"doctor could not run", Facts{CheckErr: "no git", DepsChecked: true}, "doctor"},
		{"everything at once", Facts{Checked: true, Failed: 2, Last: &Run{Command: "run lint"}, Affected: 3, Release: ready, DepsChecked: true, Outdated: 4},
			"run,setup,changes,release,updates"},
		{"a failed run comes before the failed checks", Facts{Checked: true, Failed: 1, Last: &Run{Command: "run lint"}, DepsChecked: true}, "run,setup"},
		{"a stopped run is not a failure", Facts{Checked: true, Last: &Run{Cancelled: true}, Affected: 3, DepsChecked: true}, "changes"},
		{"a run that passed is not listed", Facts{Checked: true, Last: &Run{OK: true}, DepsChecked: true}, "good"},
		{"changes before release", Facts{Checked: true, Affected: 1, Release: ready, DepsChecked: true}, "changes,release"},
		{"release alone", Facts{Checked: true, Release: ready, DepsChecked: true}, "release"},
		{"outdated dependencies", Facts{Checked: true, DepsChecked: true, Outdated: 5}, "updates"},
		{"updates not looked at come with a calm line first", Facts{Checked: true}, "good,updates"},
		{"nothing applies", Facts{Checked: true, ProjectsLoaded: true, Projects: 4, DepsChecked: true}, "good"},
		{"warnings alone are no good news", Facts{Checked: true, Warnings: 3, DepsChecked: true}, "warnings"},
		{"failed checks come before warnings", Facts{Checked: true, Failed: 1, Warnings: 2, DepsChecked: true}, "setup,warnings"},
	}
	for _, c := range cases {
		if got := kindsOf(Queue(c.f)); got != c.kinds {
			t.Errorf("%s: %s, want %s", c.name, got, c.kinds)
		}
	}
}

func TestSetupNamesTheChecksAndOffersTheFixFirst(t *testing.T) {
	f := Facts{Checked: true, Failed: 3, Fixable: 6, Problems: []string{"gradle is not on PATH", "ci.yml is out of date", "third"}, DepsChecked: true}
	it := Queue(f)[0]
	if it.Title != "3 checks fail, 6 can be fixed now" || it.Detail != "gradle is not on PATH · ci.yml is out of date" || it.Severity != SeverityFailure {
		t.Fatalf("%+v", it)
	}
	if got, want := buttonsOf(it), "Fix:ws.fix:Enter|Open Doctor:view.doctor:3"; got != want {
		t.Fatalf("buttons %q, want %q", got, want)
	}

	f.Fixable = 0
	it = Queue(f)[0]
	if it.Title != "3 checks fail" || buttonsOf(it) != "Open Doctor:view.doctor:Enter" {
		t.Fatalf("without a fix: %+v", it)
	}
	f.Failed, f.Problems = 1, nil
	if it = Queue(f)[0]; it.Title != "1 check fails" || it.Detail != "" {
		t.Fatalf("one check: %+v", it)
	}
}

func TestADoctorThatCouldNotRunOffersARetry(t *testing.T) {
	it := Queue(Facts{CheckErr: "no git"})[0]
	if it.Title != "Doctor could not run" || it.Detail != "no git" || buttonsOf(it) != "Retry:ws.doctor:Enter" {
		t.Fatalf("%+v", it)
	}
}

func TestAFailedRunNamesTheTaskAndCanBeOpenedOrRepeated(t *testing.T) {
	f := Facts{Checked: true, DepsChecked: true, Last: &Run{Command: "run lint -p smith", Task: "smith:lint", Repeatable: true}}
	it := Queue(f)[0]
	if it.Title != "smith:lint failed" || it.Kind != KindRun {
		t.Fatalf("%q", it.Title)
	}
	if got, want := buttonsOf(it), "See why:x:why-failed:Enter|Run it again:run.again:"; got != want {
		t.Fatalf("%q, want %q", got, want)
	}
	f.Last.Repeatable = false
	if got := buttonsOf(Queue(f)[0]); got != "See why:x:why-failed:Enter" {
		t.Fatalf("a command that cannot be repeated: %q", got)
	}
	f.Last.Task = ""
	if got := Queue(f)[0].Title; got != "lint -p smith failed" {
		t.Fatalf("a run that failed outside a task is named by its command: %q", got)
	}
}

func TestWarningsAreCountedAndTheFirstIsNamed(t *testing.T) {
	f := Facts{Checked: true, DepsChecked: true, Warnings: 1, Fixable: 1, WarningNotes: []string{"golangci-lint is older than CI"}}
	it := Queue(f)[0]
	if it.Title != "1 warning: golangci-lint is older than CI" || it.Severity != SeverityWarning || it.Kind != KindWarnings {
		t.Fatalf("%+v", it)
	}
	if got, want := buttonsOf(it), "Fix:ws.fix:Enter|Open Doctor:view.doctor:3"; got != want {
		t.Fatalf("%q, want %q", got, want)
	}
	f.Warnings, f.Fixable = 3, 0
	it = Queue(f)[0]
	if it.Title != "3 warnings. First: golangci-lint is older than CI" || buttonsOf(it) != "Open Doctor:view.doctor:Enter" {
		t.Fatalf("%+v", it)
	}
	for _, it := range Queue(f) {
		if it.Kind == KindGood {
			t.Fatal("All good is said while a check warns")
		}
	}
}

func TestTheCalmLineNamesTheLastRunThatPassed(t *testing.T) {
	it := Queue(Facts{Checked: true, DepsChecked: true, Last: &Run{Command: "run build", OK: true, Took: "2s"}})[0]
	if it.Kind != KindGood || it.Detail != "Last run: build passed in 2s" {
		t.Fatalf("%+v", it)
	}
}

func TestChangesNameTheProjectsAndTheTestButton(t *testing.T) {
	for n, title := range map[int]string{1: "1 project is affected by your changes", 4: "4 projects are affected by your changes"} {
		it := Queue(Facts{Checked: true, Affected: n, DepsChecked: true})[0]
		if it.Title != title || buttonsOf(it) != "Test what changed:run.affected:Enter" {
			t.Errorf("%d: %+v", n, it)
		}
	}
}

func TestTheReleaseItemNamesAPackageAndItsVersions(t *testing.T) {
	cases := []struct {
		r    Ready
		want string
	}{
		{Ready{Packages: 8, Name: "smith", From: "1.132.0", To: "1.133.0", Commits: 292}, "Ready to release: 8 packages, smith 1.132.0 → 1.133.0 (292 commits)"},
		{Ready{Packages: 1, Name: "kit", From: "0.2.0", To: "0.2.1", Commits: 1}, "Ready to release: 1 package, kit 0.2.0 → 0.2.1 (1 commit)"},
		{Ready{Packages: 2, Name: "kit", From: "0.2.0", To: "0.3.0"}, "Ready to release: 2 packages, kit 0.2.0 → 0.3.0"},
		{Ready{Packages: 1, Name: "new", To: "0.1.0"}, "Ready to release: 1 package, new 0.1.0, its first release"},
	}
	for _, c := range cases {
		r := c.r
		it := Queue(Facts{Checked: true, DepsChecked: true, Release: &r})[0]
		if it.Title != c.want || it.Kind != KindRelease || buttonsOf(it) != "Open Release:view.release:Enter" {
			t.Errorf("%+v, want %q", it, c.want)
		}
	}
}

func TestUpdatesAreReviewedWhenKnownAndLookedAtWhenNot(t *testing.T) {
	it := Queue(Facts{Checked: true, DepsChecked: true, Outdated: 3})[0]
	if it.Title != "3 dependencies have newer releases" || buttonsOf(it) != "Open Deps:view.deps:Enter" {
		t.Fatalf("%+v", it)
	}
	q := Queue(Facts{Checked: true})
	if len(q) != 2 || q[1].Title != "Updates not looked at yet" || buttonsOf(q[1]) != "Open Deps:view.deps:4" {
		t.Fatalf("%+v", q)
	}
}

func TestWhenNothingAppliesOneCalmLineOffersABuild(t *testing.T) {
	q := Queue(Facts{Checked: true, DepsChecked: true, ProjectsLoaded: true, Projects: 2})
	if len(q) != 1 || q[0].Title != "All good" || q[0].Severity != SeveritySuccess || buttonsOf(q[0]) != "Build all:run.build-all:Enter" {
		t.Fatalf("%+v", q)
	}
}

func TestAWorkspaceWithoutProjectsOffersTheTwoWaysToMakeOne(t *testing.T) {
	q := Queue(Facts{ProjectsLoaded: true})
	if len(q) != 1 || q[0].Kind != KindEmpty {
		t.Fatalf("%+v", q)
	}
	if got := buttonsOf(q[0]); got != "Add a project:file.add:a|Compose a stack:file.compose:A" {
		t.Fatalf("%q", got)
	}
	if Primary(q) != "file.add" {
		t.Fatalf("Enter runs %q", Primary(q))
	}
}

func TestWhileDoctorRunsOneItemSaysSoAndNothingCanBePressed(t *testing.T) {
	q := Queue(Facts{Waited: 5 * time.Second, Projects: 12})
	if len(q) != 1 || q[0].Kind != KindChecking || len(q[0].Buttons) != 0 || Primary(q) != "" {
		t.Fatalf("%+v", q)
	}
	if !strings.Contains(q[0].Title, "Running the health checks on the 12 projects") || !strings.Contains(q[0].Title, "5s") {
		t.Fatalf("%q", q[0].Title)
	}
}

func TestEnterRunsTheFirstButtonOfTheFirstItem(t *testing.T) {
	cases := []struct {
		name string
		f    Facts
		cmd  string
	}{
		{"fix", Facts{Checked: true, Failed: 1, Fixable: 1}, "ws.fix"},
		{"nothing to fix goes to the checks", Facts{Checked: true, Failed: 1}, "view.doctor"},
		{"retry", Facts{CheckErr: "x"}, "ws.doctor"},
		{"failed run", Facts{Checked: true, Last: &Run{}, DepsChecked: true}, "x:why-failed"},
		{"warnings", Facts{Checked: true, Warnings: 1, DepsChecked: true}, "view.doctor"},
		{"affected", Facts{Checked: true, Affected: 1, DepsChecked: true}, "run.affected"},
		{"release", Facts{Checked: true, DepsChecked: true, Release: &Ready{Packages: 1, Name: "a", To: "1.0.0"}}, "view.release"},
		{"updates", Facts{Checked: true, DepsChecked: true, Outdated: 1}, "view.deps"},
		{"good", Facts{Checked: true}, "run.build-all"},
		{"waiting", Facts{}, ""},
	}
	for _, c := range cases {
		if got := Primary(Queue(c.f)); got != c.cmd {
			t.Errorf("%s: Enter runs %q, want %q", c.name, got, c.cmd)
		}
	}
}
