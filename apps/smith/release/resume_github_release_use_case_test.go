package release

import (
	"bytes"
	"errors"
	"io"
	"strings"
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
	items, err := UnfinishedReleases(s, func(tag string) (bool, error) { return tag == "core@1.2.0", nil }, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Project != "sdk" || items[0].Tag != "sdk@0.3.0" || items[0].Next != "0.3.0" {
		t.Fatalf("items = %+v, want sdk@0.3.0 only (core has its Release, gomod and web were never released)", items)
	}
	if s.Git.TagCommit(items[0].Tag) == "" {
		t.Fatal("the item's tag must exist, so releasing it publishes to no registry")
	}
}

func TestUnfinishedReleasesIsEmptyWhenEveryTagHasItsRelease(t *testing.T) {
	items, err := UnfinishedReleases(unfinishedSession(t, true), func(string) (bool, error) { return true, nil }, io.Discard)
	if err != nil || len(items) != 0 {
		t.Fatalf("items = %+v, err = %v", items, err)
	}
}

func TestUnfinishedReleasesDoesNotAskWhenGitHubReleasesAreOff(t *testing.T) {
	asked := false
	items, err := UnfinishedReleases(unfinishedSession(t, false), func(string) (bool, error) { asked = true; return false, nil }, io.Discard)
	if err != nil || len(items) != 0 || asked {
		t.Fatalf("items = %+v, asked = %v, err = %v", items, asked, err)
	}
}

func TestUnfinishedReleasesNamesTheTagWhenTheLookupFails(t *testing.T) {
	_, err := UnfinishedReleases(unfinishedSession(t, true), func(string) (bool, error) { return false, errors.New("boom") }, io.Discard)
	if err == nil {
		t.Fatal("a failed lookup must be an error")
	}
}

func TestUnfinishedReleasesSkipsATagOlderThanHeadAndSaysSo(t *testing.T) {
	var said bytes.Buffer
	items, err := UnfinishedReleases(unfinishedSession(t, true), func(string) (bool, error) { return false, nil }, &said)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Tag != "sdk@0.3.0" {
		t.Fatalf("items = %+v, want only sdk@0.3.0, the tag at HEAD", items)
	}
	if want := "not creating the GitHub Release of core@1.2.0: HEAD is newer than its tag\n"; said.String() != want {
		t.Fatalf("said %q, want %q", said.String(), want)
	}
	if strings.Contains(said.String(), "sdk") {
		t.Fatal("the tag at HEAD must not be reported as skipped")
	}
}
