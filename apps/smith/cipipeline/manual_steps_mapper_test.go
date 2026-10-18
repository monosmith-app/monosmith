package cipipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestManualStepsNameTheFeedPermissionInAzureTerms(t *testing.T) {
	cfg := workspace.Config{
		CI:         workspace.CISection{Providers: []string{"azure"}},
		Workspace:  workspace.WorkspaceSection{Registry: "libs"},
		Registries: map[string]workspace.RegistryConfig{"libs": {Auth: "azure-artifacts", NPM: "https://pkgs.dev.azure.com/acme/_packaging/libs-feed/npm/registry/"}},
	}
	steps := strings.Join(ManualSteps(cfg, Model{}), "\n")
	for _, want := range []string{"Azure Artifacts feed libs-feed", "Feed Publisher (Contributor)", "Build Service", "upstream sources", "vsts-npm-auth"} {
		if !strings.Contains(steps, want) {
			t.Errorf("lacks %q:\n%s", want, steps)
		}
	}
	if strings.Contains(steps, "secret") {
		t.Errorf("Azure Artifacts needs no secret:\n%s", steps)
	}
}

func TestManualStepsExplainTheRemoteCacheInEachProvidersTerms(t *testing.T) {
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"azure"}}, Release: workspace.ReleaseSection{Branch: "main"}}
	cfg.Cache.Remote = "azblob://acme/smith-cache/main"
	m := cacheModel(cfg.Cache.Remote, false)
	steps := strings.Join(ManualSteps(cfg, m), "\n")
	for _, want := range []string{
		"Azure Pipelines: add the secret variables AZURE_STORAGE_SAS_TOKEN, NODE_AUTH_TOKEN, SMITH_CACHE_KEY",
		"SMITH_CACHE_KEY is any long random string",
		"never write it; only a push to main stores results",
		"SAS token scoped to it with read, write and create permission",
		"mark the variables secret",
	} {
		if !strings.Contains(steps, want) {
			t.Errorf("lacks %q:\n%s", want, steps)
		}
	}
	cfg.CI.Providers, cfg.Cache.Remote = []string{"github"}, ""
	if got := strings.Join(ManualSteps(cfg, cacheModel("", false)), "\n"); strings.Contains(got, "remote cache") {
		t.Errorf("no remote cache, no cache steps:\n%s", got)
	}
}

func TestManualStepsListSecretsButNotGitHubsOwnToken(t *testing.T) {
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github", "gitlab"}}}
	m := Model{Credentials: []protocol.Credential{{Env: "GITHUB_TOKEN"}, {Env: "FEED_TOKEN"}}}
	steps := ManualSteps(cfg, m)
	if len(steps) != 2 || !strings.Contains(steps[0], "GitHub: add the repository secrets FEED_TOKEN (") || !strings.Contains(steps[1], "GitLab: add the masked CI/CD variables GITHUB_TOKEN, FEED_TOKEN") {
		t.Errorf("steps:\n%s", strings.Join(steps, "\n"))
	}
}

func TestManualStepsWarnAProviderTheSignInDoesNotReach(t *testing.T) {
	cfg := workspace.Config{
		CI:         workspace.CISection{Providers: []string{"github"}},
		Release:    workspace.ReleaseSection{Registry: "libs"},
		Registries: map[string]workspace.RegistryConfig{"libs": {Auth: "azure-artifacts", NuGet: "https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json"}},
	}
	if steps := strings.Join(ManualSteps(cfg, Model{}), "\n"); !strings.Contains(steps, `signs in only on Azure Pipelines; for github use auth = "token"`) {
		t.Errorf("steps:\n%s", steps)
	}
}
