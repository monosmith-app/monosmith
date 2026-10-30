package release

import (
	"errors"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func unfinishedSession(t *testing.T, githubReleases bool) *session.Session {
	t.Helper()
	g := repo(t)
	commit(t, g, "packages/core/a.go", "feat: core")
	if err := g.CreateTag("core@1.2.0", "core 1.2.0"); err != nil {
		t.Fatal(err)
	}
	commit(t, g, "packages/sdk/a.go", "feat: sdk")
	if err := g.CreateTag("sdk@0.3.0", "sdk 0.3.0"); err != nil {
		t.Fatal(err)
	}
	cfg := workspace.Config{}
	cfg.Release.GitHubReleases = githubReleases
	return &session.Session{Root: g.Dir, Config: cfg, Git: g, Graph: packages(t)}
}

func TestUnfinishedReleasesFindsATagWithoutItsGitHubRelease(t *testing.T) {
	s := unfinishedSession(t, true)
	items, err := UnfinishedReleases(s, func(tag string) (bool, error) { return tag == "sdk@0.3.0", nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Project != "core" || items[0].Tag != "core@1.2.0" || items[0].Next != "1.2.0" {
		t.Fatalf("items = %+v, want core@1.2.0 only (sdk has its Release, gomod and web were never released)", items)
	}
	if s.Git.TagCommit(items[0].Tag) == "" {
		t.Fatal("the item's tag must exist, so releasing it publishes to no registry")
	}
}

func TestUnfinishedReleasesIsEmptyWhenEveryTagHasItsRelease(t *testing.T) {
	items, err := UnfinishedReleases(unfinishedSession(t, true), func(string) (bool, error) { return true, nil })
	if err != nil || len(items) != 0 {
		t.Fatalf("items = %+v, err = %v", items, err)
	}
}

func TestUnfinishedReleasesDoesNotAskWhenGitHubReleasesAreOff(t *testing.T) {
	asked := false
	items, err := UnfinishedReleases(unfinishedSession(t, false), func(string) (bool, error) { asked = true; return false, nil })
	if err != nil || len(items) != 0 || asked {
		t.Fatalf("items = %+v, asked = %v, err = %v", items, asked, err)
	}
}

func TestUnfinishedReleasesNamesTheTagWhenTheLookupFails(t *testing.T) {
	_, err := UnfinishedReleases(unfinishedSession(t, true), func(string) (bool, error) { return false, errors.New("boom") })
	if err == nil {
		t.Fatal("a failed lookup must be an error")
	}
}
