package doctor

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const billing = "The job was not started because recent account payments have failed or your spending limit needs to be increased."

func notStartedSession(t *testing.T, remote string) *session.Session {
	t.Helper()
	dir := t.TempDir()
	g := gitclient.New(dir)
	if err := g.Init(); err != nil {
		t.Skip("git is not available")
	}
	if remote != "" {
		if _, err := g.Run("remote", "add", "origin", remote); err != nil {
			t.Fatal(err)
		}
	}
	return &session.Session{Root: dir, Git: g, Config: workspace.Config{Release: workspace.ReleaseSection{Remote: "origin"}}}
}

func withFakeGitHub(t *testing.T, token, message string, err error) *int {
	t.Helper()
	oldToken, oldNote := githubToken, notStartedAnnotation
	calls := 0
	githubToken = func() string { return token }
	notStartedAnnotation = func(githubrelease.Repository, string) (string, error) { calls++; return message, err }
	t.Cleanup(func() { githubToken, notStartedAnnotation = oldToken, oldNote })
	return &calls
}

func TestABillingAnnotationIsAWarningWithTheRemedy(t *testing.T) {
	withFakeGitHub(t, "tok", billing, nil)
	got := notStartedChecks(notStartedSession(t, "https://github.com/acme/app.git"))
	if len(got) != 1 || !got[0].Warning || got[0].OK {
		t.Fatalf("findings = %+v", got)
	}
	if !strings.Contains(got[0].Detail, billing) || !strings.Contains(got[0].Remedy, "acme/app") {
		t.Errorf("finding = %+v", got[0])
	}
}

func TestNothingIsSaidWhenItCannotBeAskedOrThereIsNothing(t *testing.T) {
	cases := map[string]struct {
		remote, token, message string
		err                    error
	}{
		"no remote":     {"", "tok", billing, nil},
		"not GitHub":    {"https://gitlab.com/acme/app.git", "tok", billing, nil},
		"no login":      {"https://github.com/acme/app.git", "", billing, nil},
		"offline":       {"https://github.com/acme/app.git", "tok", "", errors.New("offline")},
		"no annotation": {"https://github.com/acme/app.git", "tok", "", nil},
	}
	for name, c := range cases {
		withFakeGitHub(t, c.token, c.message, c.err)
		if got := notStartedChecks(notStartedSession(t, c.remote)); len(got) != 0 {
			t.Errorf("%s: %+v", name, got)
		}
	}
}

func TestTheLookupCanBeSwitchedOff(t *testing.T) {
	calls := withFakeGitHub(t, "tok", billing, nil)
	t.Setenv("SMITH_NO_UPDATE_CHECK", "1")
	if got := notStartedChecks(notStartedSession(t, "https://github.com/acme/app.git")); len(got) != 0 || *calls != 0 {
		t.Errorf("findings = %+v, calls = %d", got, *calls)
	}
}
