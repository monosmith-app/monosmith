package registries

import (
	"net/url"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Markers delimit the block of .npmrc smith owns; everything outside it is
// the team's and kept as it is.
const (
	BlockStart = "# smith:registry (written from [registries] in smith.toml; smith doctor --fix rewrites it)"
	BlockEnd   = "# smith:registry-end"
)

// Block is the .npmrc block for a workspace's registries: the registry npm
// installs from, and the token each token-signed registry reads from the
// environment. GitHub Packages serves only its own scopes, so it is set for
// the packages' scope and the rest still comes from npmjs.org. Azure
// Artifacts needs no token line: npmAuthenticate@0 adds it in CI, and
// vsts-npm-auth (or artifacts-npm-credprovider) on a laptop. "" when no
// registry has an npm URL.
func Block(r *protocol.Registries, scope string) string {
	if r == nil {
		return ""
	}
	var lines []string
	if in := r.Install; in != nil && in.NPM != "" {
		if in.Auth == protocol.AuthGitHubPackages && scope != "" {
			lines = append(lines, scope+":registry="+in.NPM)
		} else {
			lines = append(lines, "registry="+in.NPM)
		}
	}
	seen := map[string]bool{}
	for _, reg := range []*protocol.Registry{r.Install, r.Publish} {
		if reg == nil || reg.NPM == "" || seen[reg.NPM] {
			continue
		}
		seen[reg.NPM] = true
		if env := TokenEnv(reg); env != "" {
			lines = append(lines, AuthPrefix(reg.NPM)+":_authToken=${"+env+"}")
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return BlockStart + "\n" + strings.Join(lines, "\n") + "\n" + BlockEnd + "\n"
}

// TokenEnv is the variable a registry's npm token is read from: the one
// smith.toml names, GITHUB_TOKEN for GitHub Packages, none for Azure
// Artifacts (its authenticate task writes the token).
func TokenEnv(r *protocol.Registry) string {
	switch r.Auth {
	case protocol.AuthToken:
		return r.TokenEnv
	case protocol.AuthGitHubPackages:
		return "GITHUB_TOKEN"
	}
	return ""
}

// AuthPrefix is how .npmrc keys a registry's credentials: the URL without
// its scheme, ending in a slash (//npm.pkg.github.com/).
func AuthPrefix(registryURL string) string {
	u, err := url.Parse(registryURL)
	if err != nil {
		return "//" + strings.TrimPrefix(strings.TrimPrefix(registryURL, "https://"), "http://")
	}
	p := u.Path
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return "//" + u.Host + p
}

// Merge replaces smith's block in an .npmrc with block, keeping every other
// line. An empty block removes smith's; a file with no block gets it on top,
// so the team's later lines still win where they set the same key.
func Merge(existing, block string) string {
	start := strings.Index(existing, BlockStart)
	if start >= 0 {
		end := strings.Index(existing[start:], BlockEnd)
		if end >= 0 {
			after := existing[start+end+len(BlockEnd):]
			after = strings.TrimPrefix(strings.TrimPrefix(after, "\r"), "\n")
			return existing[:start] + block + after
		}
	}
	if block == "" {
		return existing
	}
	return block + existing
}

// Scope is the npm scope of the workspace's packages (@acme), from the first
// scoped package name, or "".
func Scope(projects []protocol.Project) string {
	for _, p := range projects {
		for _, name := range p.Exports {
			if strings.HasPrefix(name, "@") && strings.Contains(name, "/") {
				return name[:strings.Index(name, "/")]
			}
		}
	}
	return ""
}

// Of are the registries the projects share: the host sets the same install
// registry on each, and the first project naming one decides.
func Of(projects []protocol.Project) *protocol.Registries {
	for _, p := range projects {
		if p.Registries != nil {
			return p.Registries
		}
	}
	return nil
}
