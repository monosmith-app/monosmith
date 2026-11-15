package convertplan

import (
	"regexp"
	"strings"
)

var npmrcScopedRegistry = regexp.MustCompile(`(?m)^[ \t]*(@[\w.-]+):registry[ \t]*=[ \t]*(\S+)`)

// azureArtifactsNPM matches an Azure Artifacts feed's npm registry URL.
var azureArtifactsNPM = regexp.MustCompile(`^https://pkgs\.dev\.azure\.com/[^/]+(?:/[^/]+)?/_packaging/([^/]+)/npm/registry/?$`)

// NpmrcFeed finds the Azure Artifacts feed a workspace's .npmrc routes a
// scope to (@auto:registry=https://pkgs.dev.azure.com/org/_packaging/AUTO/npm/registry/),
// which is where its packages publish: the feed's name for smith.toml's
// [registries], and its URL. A comment line names nothing.
func NpmrcFeed(npmrc string) (name, url string, ok bool) {
	for _, m := range npmrcScopedRegistry.FindAllStringSubmatch(npmrc, -1) {
		if feed := azureArtifactsNPM.FindStringSubmatch(m[2]); feed != nil {
			return registryName(feed[1]), m[2], true
		}
	}
	return "", "", false
}

// registryName is a feed's name as a [registries] key: lower case, and
// letters, digits and hyphens only.
func registryName(feed string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(feed) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		case r == '.', r == '_':
			b.WriteRune('-')
		}
	}
	return b.String()
}
