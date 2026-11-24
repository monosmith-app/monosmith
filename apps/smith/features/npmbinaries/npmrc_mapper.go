package npmbinaries

import (
	"net/url"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PublicRegistry is the registry the packages publish to when smith.toml names
// none.
const PublicRegistry = "https://registry.npmjs.org/"

// RegistryOf is where an app's npm packages publish to, from the registry
// smith.toml names for the project ([projects.<name>] registry, else
// [release] registry, else [workspace] registry), and the variable npm reads
// its token from: GITHUB_TOKEN for GitHub Packages, the named variable for a
// token registry, SYSTEM_ACCESSTOKEN for Azure Artifacts (whose
// npmAuthenticate@0 task usually writes the token into the workspace's .npmrc
// instead), NODE_AUTH_TOKEN for npmjs.org. A nil registry, or one with no npm
// URL, is npmjs.org.
func RegistryOf(reg *protocol.Registry) (registryURL, tokenEnv string) {
	if reg == nil || reg.NPM == "" {
		return "", "NODE_AUTH_TOKEN"
	}
	switch reg.Auth {
	case protocol.AuthAzureArtifacts:
		return reg.NPM, "SYSTEM_ACCESSTOKEN"
	case protocol.AuthGitHubPackages:
		return reg.NPM, "GITHUB_TOKEN"
	case protocol.AuthToken:
		return reg.NPM, reg.TokenEnv
	}
	return reg.NPM, "NODE_AUTH_TOKEN"
}

// Npmrc is the .npmrc a package folder publishes with. npm reads a project
// .npmrc only from the folder that holds the package.json, so the workspace's
// .npmrc does not reach the generated packages: Npmrc names the release's
// registry, copies the workspace's credential lines for it (a named
// registry only; the ones Azure's npmAuthenticate@0 writes, say), and adds a
// line reading r.TokenEnv when env has it, last so it wins. "" when there is nothing to say: npm's own login
// and registry are used.
//
// Parameters: workspace is the text of the workspace's .npmrc ("" for none);
// r carries Registry and TokenEnv ("" on npmjs.org is NODE_AUTH_TOKEN); env
// reads the environment.
func Npmrc(workspace string, r Release, env func(string) string) string {
	registry := r.Registry
	if registry == "" {
		registry = PublicRegistry
	}
	prefix := AuthPrefix(registry)
	var lines []string
	if r.Registry != "" {
		lines = append(lines, "registry="+r.Registry)
	} else {
		// npmjs.org keeps npm's own login: a copied ${VAR} line whose
		// variable is unset would make npm refuse to start.
		workspace = ""
	}
	for _, line := range strings.Split(strings.ReplaceAll(workspace, "\r\n", "\n"), "\n") {
		key, _, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || !strings.HasPrefix(key, "//") {
			continue
		}
		at := strings.LastIndex(key, ":")
		if at < 0 || !strings.HasPrefix(prefix, key[:at]) {
			continue
		}
		lines = append(lines, strings.TrimSpace(line))
	}
	tokenEnv := r.TokenEnv
	if tokenEnv == "" && r.Registry == "" {
		tokenEnv = "NODE_AUTH_TOKEN"
	}
	if tokenEnv != "" && env(tokenEnv) != "" {
		lines = append(lines, prefix+":_authToken=${"+tokenEnv+"}")
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

// AuthPrefix is how .npmrc keys a registry's credentials: the URL without its
// scheme, ending in a slash (//pkgs.dev.azure.com/acme/_packaging/auto/npm/registry/).
func AuthPrefix(registryURL string) string {
	u, err := url.Parse(registryURL)
	if err != nil || u.Host == "" {
		return "//" + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(registryURL, "https://"), "http://"), "/") + "/"
	}
	return "//" + u.Host + strings.TrimSuffix(u.Path, "/") + "/"
}
