package npmbinaries

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

const azureFeed = "https://pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/"

func envOf(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestRegistryOfFollowsTheSignInSmithTomlNames(t *testing.T) {
	cases := []struct {
		name          string
		reg           *protocol.Registry
		wantURL, want string
	}{
		{"public", nil, "", "NODE_AUTH_TOKEN"},
		{"no npm URL", &protocol.Registry{Auth: protocol.AuthAzureArtifacts, NuGet: "https://x"}, "", "NODE_AUTH_TOKEN"},
		{"azure", &protocol.Registry{Auth: protocol.AuthAzureArtifacts, NPM: azureFeed}, azureFeed, "SYSTEM_ACCESSTOKEN"},
		{"github", &protocol.Registry{Auth: protocol.AuthGitHubPackages, NPM: "https://npm.pkg.github.com"}, "https://npm.pkg.github.com", "GITHUB_TOKEN"},
		{"token", &protocol.Registry{Auth: protocol.AuthToken, TokenEnv: "FEED_TOKEN", NPM: "https://npm.acme.dev"}, "https://npm.acme.dev", "FEED_TOKEN"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url, env := RegistryOf(c.reg)
			if url != c.wantURL || env != c.want {
				t.Errorf("RegistryOf = %q, %q; want %q, %q", url, env, c.wantURL, c.want)
			}
		})
	}
}

func TestNpmrc(t *testing.T) {
	workspace := strings.Join([]string{
		"registry=" + azureFeed,
		"always-auth=true",
		"//pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/:_authToken=from-npmAuthenticate",
		"//pkgs.dev.azure.com/acme/_packaging/auto/npm/:_authToken=from-npmAuthenticate",
		"//pkgs.dev.azure.com/acme/_packaging/other/npm/registry/:_authToken=other-feed",
		"//registry.npmjs.org/:_authToken=${NPM_TOKEN}",
	}, "\r\n")

	t.Run("a named registry is named and keeps the workspace's sign-in for it", func(t *testing.T) {
		got := Npmrc(workspace, Release{Registry: azureFeed, TokenEnv: "SYSTEM_ACCESSTOKEN"}, envOf(nil))
		want := "registry=" + azureFeed + "\n" +
			"//pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/:_authToken=from-npmAuthenticate\n" +
			"//pkgs.dev.azure.com/acme/_packaging/auto/npm/:_authToken=from-npmAuthenticate\n"
		if got != want {
			t.Errorf("Npmrc =\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("the registry's token variable comes last so it wins", func(t *testing.T) {
		got := Npmrc("", Release{Registry: azureFeed, TokenEnv: "SYSTEM_ACCESSTOKEN"}, envOf(map[string]string{"SYSTEM_ACCESSTOKEN": "job-token"}))
		want := "registry=" + azureFeed + "\n//pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/:_authToken=${SYSTEM_ACCESSTOKEN}\n"
		if got != want {
			t.Errorf("Npmrc =\n%s\nwant\n%s", got, want)
		}
	})

	t.Run("npmjs.org reads NODE_AUTH_TOKEN and copies nothing", func(t *testing.T) {
		got := Npmrc(workspace, Release{}, envOf(map[string]string{"NODE_AUTH_TOKEN": "t"}))
		if got != "//registry.npmjs.org/:_authToken=${NODE_AUTH_TOKEN}\n" {
			t.Errorf("Npmrc = %q", got)
		}
	})

	t.Run("npmjs.org without a token leaves npm's own login", func(t *testing.T) {
		if got := Npmrc(workspace, Release{}, envOf(nil)); got != "" {
			t.Errorf("Npmrc = %q", got)
		}
	})
}
