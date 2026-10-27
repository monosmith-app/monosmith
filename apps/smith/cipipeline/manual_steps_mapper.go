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
		if p == "circleci" {
			steps = append(steps, "CircleCI: the checkout key is read-only, so add a user key with write access (Project Settings → SSH Keys → Add user key) for smith ci release to push tags; for the release preview comment on GitHub pull requests add GITHUB_TOKEN too")
		}
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
		case "circleci":
			steps = append(steps, "CircleCI: add the environment variables "+strings.Join(secrets, ", ")+" (Project Settings → Environment Variables, or a context the project uses)")
		case "gitlab":
			steps = append(steps, "GitLab: add the masked CI/CD variables "+strings.Join(secrets, ", ")+" (Settings → CI/CD → Variables)")
		}
	}
	steps = append(steps, cacheSteps(cfg)...)
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
			if r.Maven != "" {
				steps = append(steps, fmt.Sprintf("registry %s (GitHub Packages, Maven): anyone who depends on these packages needs their own token with read:packages in ~/.m2/settings.xml or Gradle's credentials, even for a public repository; GitHub Packages serves no Maven package anonymously", name))
			}
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

// cacheSteps are what a person does by hand for the remote cache: the
// secrets are listed with the others; these say what each one is and how it
// behaves where no secret exists.
func cacheSteps(cfg workspace.Config) []string {
	remote := cfg.Cache.Remote
	if remote == "" {
		return nil
	}
	steps := []string{
		"remote cache " + remote + ": SMITH_CACHE_KEY is any long random string the whole team shares; every object is signed with it, and with no value (a pull request from a fork, which gets no secrets) the remote is off and the run uses the local cache",
		"remote cache: pull requests and merge requests read the cache and never write it; only a push to " + cfg.Release.Branch + " stores results",
	}
	switch {
	case strings.HasPrefix(remote, "azblob://"):
		steps = append(steps, "remote cache (Azure Blob): create the container, then a SAS token scoped to it with read, write and create permission, and keep it as AZURE_STORAGE_SAS_TOKEN; a developer who only reads needs read permission alone")
	case strings.HasPrefix(remote, "s3://"):
		steps = append(steps, "remote cache (S3): AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY belong to an identity allowed to read and write the bucket; set AWS_REGION to its region")
	case strings.HasPrefix(remote, "https://"), strings.HasPrefix(remote, "http://"):
		steps = append(steps, "remote cache (HTTP): the server must take PUT, GET and HEAD; if it needs a bearer token, keep it as SMITH_CACHE_TOKEN")
	}
	if slices.Contains(cfg.CI.Providers, "azure") {
		steps = append(steps, "remote cache, on Azure Pipelines: mark the variables secret; the pipeline maps each into the steps that run tasks, since Azure exposes a secret only through env:. In a build of a fork Azure leaves the variable's name in place of its value, which smith reads as unset")
	}
	return steps
}

// azureFeed is the feed's name in an Azure Artifacts URL (the segment after
// _packaging), or the registry's first URL when it has none.
func azureFeed(r workspace.RegistryConfig) string {
	for _, raw := range []string{r.NPM, r.PyPI, r.PyPIUpload, r.NuGet, r.Maven} {
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
