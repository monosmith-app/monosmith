package release

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/apps/smith/splitrepo"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Preflight refuses a real release that would fail halfway or corrupt history:
// a shallow clone, the wrong branch, uncommitted changes (manifests are
// stamped and restored), or a missing publish credential.
func Preflight(s *session.Session, plan Plan) error {
	var problems []string
	if s.Git.IsShallow() {
		problems = append(problems, "the clone is shallow, so versions cannot be computed from history: fetch with full depth (actions/checkout fetch-depth: 0)")
	}
	if branch := s.Git.CurrentBranch(); branch != "" && branch != s.Config.Release.Branch && os.Getenv("SMITH_RELEASE_ANY_BRANCH") == "" {
		problems = append(problems, fmt.Sprintf("releases run from %s, not %s", s.Config.Release.Branch, branch))
	}
	if out, err := s.Git.Run("status", "--porcelain", "--untracked-files=no"); err == nil {
		if changed := Uncommitted(out); len(changed) > 0 {
			problems = append(problems, "the working tree has uncommitted changes ("+strings.Join(changed, ", ")+"); commit or stash them first")
		}
	}
	if s.Config.Release.GitHubReleases && len(plan.Items) > 0 {
		if _, err := githubrelease.FromEnvironment(); err != nil {
			problems = append(problems, err.Error())
		}
		if _, err := Repository(s); err != nil {
			problems = append(problems, err.Error())
		}
	}
	plugins := map[string]bool{}
	mirrored := false
	for _, item := range plan.Items {
		plugins[item.Plugin] = true
		// A deploy plugin (docker) publishes the project too, so its
		// credentials are checked as well.
		if p := s.Graph.Get(item.Project); p != nil {
			for _, d := range p.Deployers {
				plugins[d] = true
			}
		}
		if p := s.Graph.Get(item.Project); p != nil && !Native(p) {
			for _, c := range DistributionCredentials(p.Release) {
				if os.Getenv(c.Env) == "" && !alternative(c.Env) {
					problems = append(problems, fmt.Sprintf("%s is not set: %s", c.Env, c.Description))
				}
			}
		}
		if p := s.Graph.Get(item.Project); p != nil && p.SplitRepo != "" {
			if splitrepo.Token() == "" {
				c := splitrepo.Credential(p.SplitRepo)
				problems = append(problems, fmt.Sprintf("%s is not set: %s", c.Env, c.Description))
			}
			mirrored = true
		}
	}
	if mirrored {
		problems = append(problems, splitrepo.Ready(s.Git)...)
	}
	for name := range plugins {
		var projects []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.ServedBy(name) {
				projects = append(projects, p.Project)
			}
		}
		var ci protocol.CIResult
		ok, err := s.Plugins.Call(name, protocol.MethodCI, protocol.CIParams{Projects: projects}, &ci, time.Minute)
		if err != nil {
			// A plugin that cannot say what it needs (docker, with registry
			// settings that contradict each other) would fail the release
			// later, half-way: say so now.
			problems = append(problems, fmt.Sprintf("%s: %v", name, err))
			continue
		}
		if !ok {
			continue
		}
		for _, c := range ci.Credentials {
			if c.Required && os.Getenv(c.Env) == "" && !alternative(c.Env) {
				problems = append(problems, fmt.Sprintf("%s is not set: %s", c.Env, c.Description))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("cannot release:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

// DistributionCredentials are the secrets publishing an app's binaries to
// npm, PyPI, NuGet and RubyGems needs.
func DistributionCredentials(r *workspace.ProjectRelease) []protocol.Credential {
	var creds []protocol.Credential
	if r == nil {
		return nil
	}
	if len(r.NPM) > 0 {
		creds = append(creds, protocol.Credential{Env: "NODE_AUTH_TOKEN", Description: "npm token that can publish " + strings.Join(r.NPM, " and ") + " and their platform packages", Required: true})
	}
	if r.PyPI != "" {
		creds = append(creds, protocol.Credential{Env: "PYPI_TOKEN", Description: "PyPI API token that can upload " + r.PyPI, Required: true})
	}
	if r.NuGet != "" {
		creds = append(creds, protocol.Credential{Env: "NUGET_API_KEY", Description: "NuGet API key that can push " + r.NuGet + " and " + r.NuGet + ".*", Required: true})
	}
	if r.RubyGems != "" {
		creds = append(creds, protocol.Credential{Env: "GEM_HOST_API_KEY", Description: "RubyGems API key that can push " + r.RubyGems, Required: true})
	}
	return creds
}

// alternative accepts the credential forms each tool also understands.
func alternative(env string) bool {
	switch env {
	case "PYPI_TOKEN":
		return os.Getenv("TWINE_PASSWORD") != ""
	case "NODE_AUTH_TOKEN":
		return os.Getenv("NPM_TOKEN") != "" || os.Getenv("SMITH_NPM_ALREADY_AUTHENTICATED") != ""
	}
	return false
}

// Uncommitted lists the changed files of `git status --porcelain` that
// matter to a release. go.work.sum is left out: it caches module checksums,
// `go mod download` (CI setup) completes it, and it never changes what is
// released.
func Uncommitted(porcelain string) []string {
	var out []string
	for _, line := range strings.Split(strings.ReplaceAll(porcelain, "\r\n", "\n"), "\n") {
		if len(line) < 4 {
			continue
		}
		file := strings.TrimSpace(line[3:])
		if file == "go.work.sum" {
			continue
		}
		out = append(out, file)
	}
	return out
}
