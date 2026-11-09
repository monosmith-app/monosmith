package release

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

func TestStableOnlyLeavesThePyPITokenOutOfAPrereleaseButKeepsItForTheWorkflow(t *testing.T) {
	r := &workspace.ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi"}}
	if got := DistributionCredentials(r, "2.0.0-beta.1"); len(got) != 0 {
		t.Fatalf("a pre-release needs no PyPI token: %+v", got)
	}
	if got := DistributionCredentials(r, "2.0.0"); len(got) != 1 {
		t.Fatalf("a stable version needs it: %+v", got)
	}
	if got := DistributionCredentials(r, ""); len(got) != 1 || got[0].Description == "PyPI API token that can upload acme" {
		t.Fatalf("the workflow keeps the secret and says it is for stable releases: %+v", got)
	}
}

func TestTheRubyGemsKeyIsSignedInByTheRunOnlyWithTrustedPublishing(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.RubyGemsTrustedPublishing = true
	trusted := &session.Session{Config: cfg}
	cases := []struct {
		s    *session.Session
		env  string
		want bool
	}{
		{secret, "GEM_HOST_API_KEY", false},
		{trusted, "GEM_HOST_API_KEY", true},
		{trusted, "NODE_AUTH_TOKEN", false},
	}
	for _, c := range cases {
		if got := signedInByTheRun(c.s, c.env); got != c.want {
			t.Errorf("signedInByTheRun(trusted=%v, %s) = %v, want %v", c.s.Config.Release.RubyGemsTrustedPublishing, c.env, got, c.want)
		}
	}
}

func TestTheMarketplaceTokenIsSignedInByTheRunOnlyWithEntra(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.VSCEEntra = true
	entra := &session.Session{Config: cfg}
	for _, c := range []struct {
		s    *session.Session
		env  string
		want bool
	}{
		{secret, "VSCE_PAT", false},
		{entra, "VSCE_PAT", true},
		{entra, "NODE_AUTH_TOKEN", false},
	} {
		if got := signedInByTheRun(c.s, c.env); got != c.want {
			t.Errorf("signedInByTheRun(entra=%v, %s) = %v, want %v", c.s.Config.Release.VSCEEntra, c.env, got, c.want)
		}
	}
}
