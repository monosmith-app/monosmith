package doctor

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/recipereleases"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const pinned = "https://github.com/acme/recipes/releases/download/team@1.2.0/team.recipe.toml"

func releaseSession(recipe string, byKind map[string]string) *session.Session {
	return &session.Session{Config: workspace.Config{Workspace: workspace.WorkspaceSection{Recipe: recipe}, Recipes: byKind}}
}

func withLatest(t *testing.T, version string, err error) {
	t.Helper()
	original := latestRelease
	latestRelease = func(recipereleases.Location) (string, error) { return version, err }
	t.Cleanup(func() { latestRelease = original })
}

func TestANewerReleaseOfAPinnedRecipeIsAWarningWithTheNewUrl(t *testing.T) {
	withLatest(t, "1.10.0", nil)
	got := recipeReleaseChecks(releaseSession(pinned, nil))
	if len(got) != 1 || !got[0].Warning || got[0].OK {
		t.Fatalf("findings = %+v", got)
	}
	want := "https://github.com/acme/recipes/releases/download/team@1.10.0/team.recipe.toml"
	if !strings.Contains(got[0].Remedy, want) || !strings.Contains(got[0].Detail, "team 1.10.0 is released") {
		t.Errorf("finding = %+v", got[0])
	}
}

func TestTheNewestReleaseIsOk(t *testing.T) {
	withLatest(t, "1.2.0", nil)
	got := recipeReleaseChecks(releaseSession(pinned, map[string]string{"app": pinned}))
	if len(got) != 1 || !got[0].OK {
		t.Errorf("findings = %+v; one recipe pinned twice is one finding", got)
	}
}

func TestRecipesThatAreNotReleasePinnedOrCannotBeAskedAboutSayNothing(t *testing.T) {
	withLatest(t, "9.0.0", nil)
	for _, ref := range []string{"", "./team.recipe.toml", "https://github.com/acme/recipes/raw/main/team.recipe.toml"} {
		if got := recipeReleaseChecks(releaseSession(ref, nil)); len(got) != 0 {
			t.Errorf("%q: %+v", ref, got)
		}
	}
	withLatest(t, "", errors.New("offline"))
	if got := recipeReleaseChecks(releaseSession(pinned, nil)); len(got) != 0 {
		t.Errorf("offline: %+v", got)
	}
	withLatest(t, "9.0.0", nil)
	t.Setenv("SMITH_NO_UPDATE_CHECK", "1")
	if got := recipeReleaseChecks(releaseSession(pinned, nil)); len(got) != 0 {
		t.Errorf("SMITH_NO_UPDATE_CHECK: %+v", got)
	}
}
