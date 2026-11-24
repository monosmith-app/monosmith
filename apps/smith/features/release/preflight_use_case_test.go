package release

import (
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestTheNpmSecretFollowsTheRegistrySmithTomlNames(t *testing.T) {
	r := &workspace.ProjectRelease{NPM: []string{"@auto/jazure-cli"}}
	cases := []struct {
		name string
		reg  *protocol.Registry
		want []string
	}{
		{"public", nil, []string{"NODE_AUTH_TOKEN"}},
		{"azure-artifacts signs in with the pipeline's own token", &protocol.Registry{Auth: protocol.AuthAzureArtifacts, NPM: "https://pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/"}, nil},
		{"github-packages", &protocol.Registry{Auth: protocol.AuthGitHubPackages, NPM: "https://npm.pkg.github.com"}, []string{"GITHUB_TOKEN"}},
		{"token", &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "FEED_TOKEN", NPM: "https://npm.acme.dev"}, []string{"FEED_TOKEN"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []string
			for _, cred := range DistributionCredentials(r, c.reg, "") {
				got = append(got, cred.Env)
			}
			if len(got) != len(c.want) || (len(got) == 1 && got[0] != c.want[0]) {
				t.Errorf("credentials = %v, want %v", got, c.want)
			}
		})
	}
}

func TestStableOnlyLeavesThePyPITokenOutOfAPrereleaseButKeepsItForTheWorkflow(t *testing.T) {
	r := &workspace.ProjectRelease{PyPI: "acme", StableOnly: []string{"pypi"}}
	if got := DistributionCredentials(r, nil, "2.0.0-beta.1"); len(got) != 0 {
		t.Fatalf("a pre-release needs no PyPI token: %+v", got)
	}
	if got := DistributionCredentials(r, nil, "2.0.0"); len(got) != 1 {
		t.Fatalf("a stable version needs it: %+v", got)
	}
	if got := DistributionCredentials(r, nil, ""); len(got) != 1 || got[0].Description == "PyPI API token that can upload acme" {
		t.Fatalf("the workflow keeps the secret and says it is for stable releases: %+v", got)
	}
}

func TestTheRubyGemsKeyIsSignedInByTheRunOnlyWithTrustedPublishing(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.RubyGemsTrustedPublishing = new(true)
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
			t.Errorf("signedInByTheRun(trusted=%v, %s) = %v, want %v", c.s.Config.Release.RubyGemsTrustedOn(), c.env, got, c.want)
		}
	}
}

func TestTheMarketplaceTokenIsSignedInByTheRunOnlyWithEntra(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.VSCEEntra = new(true)
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
			t.Errorf("signedInByTheRun(entra=%v, %s) = %v, want %v", c.s.Config.Release.VSCEEntraOn(), c.env, got, c.want)
		}
	}
}

func TestTheNuGetKeyIsSignedInByTheRunOnlyWithTrustedPublishing(t *testing.T) {
	cfg := workspace.Config{}
	secret := &session.Session{Config: cfg}
	cfg.Release.NuGetTrustedPublishing = new(true)
	trusted := &session.Session{Config: cfg}
	for _, c := range []struct {
		s    *session.Session
		env  string
		want bool
	}{
		{secret, "NUGET_API_KEY", false},
		{trusted, "NUGET_API_KEY", true},
		{trusted, "GEM_HOST_API_KEY", false},
	} {
		if got := signedInByTheRun(c.s, c.env); got != c.want {
			t.Errorf("signedInByTheRun(trusted=%v, %s) = %v, want %v", c.s.Config.Release.NuGetTrustedOn(), c.env, got, c.want)
		}
	}
}
