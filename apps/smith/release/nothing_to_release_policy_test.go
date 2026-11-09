package release

import (
	"strings"
	"testing"
)

func TestNothingToReleaseExplainsWhichCommitsCountAndShowsTheFirstReleaseCommand(t *testing.T) {
	text := NothingToRelease(Plan{Unchanged: []string{"sdk"}})
	for _, want := range []string{"feat:", "fix:", "chore: add app web", "never do", "touches its own folder", "only under tools/ or in smith.toml, releases nothing", FirstReleaseCommand, "smith does not make it"} {
		if !strings.Contains(text, want) {
			t.Errorf("the explanation lacks %q:\n%s", want, text)
		}
	}
}

func TestAWorkspaceWithoutAPackageSaysAppsAndLibsNeverReleaseAndHasNoEmptyCommit(t *testing.T) {
	text := NothingToRelease(Plan{})
	for _, want := range []string{"no package", "Apps and internal libraries are never released", "smith add package", "smith kind <name> package"} {
		if !strings.Contains(text, want) {
			t.Errorf("the hint lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "--allow-empty") || strings.Contains(text, "first version") {
		t.Errorf("an empty feat: commit does nothing without a package:\n%s", text)
	}
}

func TestAHeldPackageIsNamedAndTheEmptyCommitIsNotSuggested(t *testing.T) {
	plan, notes, err := Adjust(Plan{Items: []Item{{Project: "sdk", Next: "0.1.0", Reason: "1 releasing commit(s)"}}}, Adjustments{Hold: []string{"sdk"}})
	if err != nil || len(notes) == 0 {
		t.Fatalf("%v %v", notes, err)
	}
	text := NothingToRelease(plan)
	if !strings.Contains(text, "you held back sdk") || !strings.Contains(text, "without --hold") || strings.Contains(text, "--allow-empty") {
		t.Errorf("got:\n%s", text)
	}
}

func TestTheBumpOfANeverReleasedPackageExplainsTheVersionCommit(t *testing.T) {
	_, _, err := Adjust(Plan{Items: []Item{{Project: "sdk", Next: "0.1.0", Level: "minor"}}}, Adjustments{Bump: map[string]string{"sdk": "major"}})
	if err == nil {
		t.Fatal("no error")
	}
	for _, want := range []string{"version(sdk)[x.y.z]", `git commit --allow-empty -m "version(sdk)[1.0.0]"`, "square brackets"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the message lacks %q: %v", want, err)
		}
	}
}
