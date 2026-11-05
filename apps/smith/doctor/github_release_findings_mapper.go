package doctor

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/releasepack"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// publicCredential is the variable each ecosystem's public registry publish
// reads its credential from.
var publicCredential = map[string]string{
	"node": "NODE_AUTH_TOKEN", "python": "PYPI_TOKEN", "dotnet": "NUGET_API_KEY", "ruby": "GEM_HOST_API_KEY", "rust": "CARGO_REGISTRY_TOKEN",
}

// githubReleaseChecks reports the registry = "github-release" setting: which
// packages attach to GitHub Releases, whether their plugin can pack, and,
// when the workspace uses the mode, which other package would publish to a
// public registry with a credential that is not set (a private repository
// cannot publish there). Nothing is reported for a workspace that does not
// use the mode.
func githubReleaseChecks(s *session.Session) []protocol.Finding {
	return githubReleaseFindings(s.Graph.Projects, s.Config.Release.GitHubReleases, func(plugin string) bool {
		c := s.Plugins.Get(plugin)
		return c != nil && c.Info.Capabilities.Packs
	}, os.Getenv)
}

// githubReleaseFindings is githubReleaseChecks with its inputs given.
func githubReleaseFindings(all []*projectgraph.Project, githubReleases bool, packs func(plugin string) bool, env func(string) string) []protocol.Finding {
	projects := map[string]*projectgraph.Project{}
	names := make([]string, 0, len(all))
	for _, p := range all {
		projects[p.Name] = p
		names = append(names, p.Name)
	}
	sort.Strings(names)
	var attached []string
	var out []protocol.Finding
	for _, name := range names {
		p := projects[name]
		if !release.PacksToRelease(p) {
			continue
		}
		attached = append(attached, name)
		if _, ok := releasepack.EcosystemOf(p.Plugin); ok && !packs(p.Plugin) {
			out = append(out, protocol.Finding{Check: "release: " + name + " can be packed", Detail: fmt.Sprintf("registry = \"github-release\", but the %s plugin cannot pack a package", p.Plugin), Remedy: "update the plugin; until then the release refuses to run, so nothing is published to a public registry"})
		}
	}
	if len(attached) == 0 {
		return out
	}
	out = append(out, protocol.Finding{Check: "release: github_releases is on for the packages that attach to GitHub Releases", OK: githubReleases,
		Detail: "registry = \"github-release\" attaches package files to GitHub Releases, which are off", Remedy: "set github_releases = true under [release] in smith.toml"})
	out = append(out, protocol.Finding{Check: fmt.Sprintf("release: %d package(s) attach to GitHub Releases, not to a registry: %s", len(attached), strings.Join(attached, ", ")), OK: true})
	for _, name := range names {
		p := projects[name]
		if p.Kind != protocol.ProjectPackage || release.Native(p) || release.PublishesToNoRegistry(p) || !p.Publishes() {
			continue
		}
		if p.Registries != nil && p.Registries.Publish != nil {
			continue // a private registry of its own
		}
		credential, ok := publicCredential[p.Plugin]
		if !ok || env(credential) != "" {
			continue
		}
		out = append(out, protocol.Finding{Check: "release: " + name + " has the credential its public registry asks for", Warning: true,
			Detail: fmt.Sprintf("%s would publish to the public %s registry, and %s is not set; a private repository must not publish there", name, p.Plugin, credential),
			Remedy: fmt.Sprintf("set [projects.%s] registry = \"github-release\" to attach its package file to the GitHub Release, or \"none\" to release it by its tag only", name)})
	}
	return out
}
