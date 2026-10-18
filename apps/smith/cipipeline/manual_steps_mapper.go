package cipipeline

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ManualSteps are what a person must set up by hand before the generated
// pipelines work, in each provider's terms: the secrets to create, and for
// each private registry the workspace uses, the permission or sign-in it
// needs.
func ManualSteps(cfg workspace.Config, m Model) []string {
	var steps []string
	for _, p := range cfg.CI.Providers {
		var secrets []string
		for _, c := range m.Credentials {
			if p == "github" && c.Env == "GITHUB_TOKEN" {
				continue // the workflow's own token
			}
			secrets = append(secrets, c.Env)
		}
		if len(secrets) == 0 {
			continue
		}
		switch p {
		case "github":
			steps = append(steps, "GitHub: add the repository secrets "+strings.Join(secrets, ", ")+" (Settings → Secrets and variables → Actions)")
		case "azure":
			steps = append(steps, "Azure Pipelines: add the secret variables "+strings.Join(secrets, ", ")+" to the pipeline (Edit → Variables, keep this value secret)")
		case "gitlab":
			steps = append(steps, "GitLab: add the masked CI/CD variables "+strings.Join(secrets, ", ")+" (Settings → CI/CD → Variables)")
		}
	}
	for _, name := range usedRegistries(cfg) {
		r := cfg.Registries[name]
		switch r.Auth {
		case "azure-artifacts":
			feed := azureFeed(r)
			steps = append(steps,
				fmt.Sprintf("registry %s (Azure Artifacts feed %s): in the feed's settings → Permissions, give the project's build service (\"<project> Build Service (<organization>)\") the Feed Publisher (Contributor) role, so the pipeline's own token can install and publish", name, feed),
				fmt.Sprintf("registry %s: add upstream sources to the feed (npmjs.com, PyPI, nuget.org) so public packages install through it", name),
				fmt.Sprintf("registry %s, on a laptop: sign in once with vsts-npm-auth (npm), artifacts-keyring (pip) or the Azure Artifacts Credential Provider (dotnet)", name))
			if !slices.Contains(cfg.CI.Providers, "azure") {
				steps = append(steps, fmt.Sprintf("registry %s: auth \"azure-artifacts\" signs in only on Azure Pipelines; for %s use auth = \"token\" with a personal access token that can read and write packages", name, strings.Join(cfg.CI.Providers, " and ")))
			}
		case "github-packages":
			steps = append(steps, fmt.Sprintf("registry %s (GitHub Packages): the workflow's GITHUB_TOKEN reads and writes this repository's packages; to install a package another repository publishes, give this repository access in that package's settings (Manage Actions access)", name))
			if !slices.Contains(cfg.CI.Providers, "github") {
				steps = append(steps, fmt.Sprintf("registry %s: GITHUB_TOKEN exists only in GitHub Actions; elsewhere use auth = \"token\" with a personal access token that has read:packages and write:packages", name))
			}
		case "token":
			steps = append(steps, fmt.Sprintf("registry %s: set %s locally too, so installs on a laptop can sign in", name, r.TokenEnv))
		}
	}
	return steps
}

// usedRegistries are the registries smith.toml refers to, sorted.
func usedRegistries(cfg workspace.Config) []string {
	used := map[string]bool{}
	for _, n := range []string{cfg.Workspace.Registry, cfg.Release.Registry} {
		if n != "" {
			used[n] = true
		}
	}
	for _, po := range cfg.Projects {
		if po.Registry != "" {
			used[po.Registry] = true
		}
	}
	out := make([]string, 0, len(used))
	for n := range used {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// azureFeed is the feed's name in an Azure Artifacts URL (the segment after
// _packaging), or the registry's first URL when it has none.
func azureFeed(r workspace.RegistryConfig) string {
	for _, raw := range []string{r.NPM, r.PyPI, r.PyPIUpload, r.NuGet} {
		u, err := url.Parse(raw)
		if raw == "" || err != nil {
			continue
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i, p := range parts {
			if p == "_packaging" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
		return raw
	}
	return ""
}
