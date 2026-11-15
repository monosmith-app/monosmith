package registrypage

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	npmrcRegistry = regexp.MustCompile(`(?m)^\s*(?:@[^:\s]+:)?registry\s*=\s*(\S+)`)
	nugetSource   = regexp.MustCompile(`value\s*=\s*"(https?://[^"]+)"`)
)

// AmbientPrivate reports whether the workspace points the plugin's tool at a
// registry other than the public one without smith.toml saying so: a registry
// line in the root .npmrc, PIP_INDEX_URL, or a source in a NuGet.Config. The
// page of a package is then not known, so the caller shows none.
//
// Parameters: plugin is the language plugin; root is the workspace root.
// Returns: true when a private registry is in use.
func AmbientPrivate(plugin, root string) bool {
	switch plugin {
	case "node":
		return anyPrivate(npmrcRegistry, filepath.Join(root, ".npmrc"))
	case "python":
		raw := os.Getenv("PIP_INDEX_URL")
		return raw != "" && !isPublic(raw)
	case "dotnet":
		for _, n := range []string{"NuGet.Config", "nuget.config", "NuGet.config"} {
			if anyPrivate(nugetSource, filepath.Join(root, n)) {
				return true
			}
		}
	}
	return false
}

// anyPrivate reads a file and reports whether a registry address in it, found
// by pattern, is not public. A missing file has none.
func anyPrivate(pattern *regexp.Regexp, file string) bool {
	raw, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	for _, m := range pattern.FindAllStringSubmatch(string(raw), -1) {
		if !isPublic(m[1]) {
			return true
		}
	}
	return false
}

// isPublic reports whether an address is on a public registry host.
func isPublic(raw string) bool {
	u, err := url.Parse(strings.Trim(raw, `"'`))
	return err == nil && publicHosts[strings.ToLower(u.Hostname())]
}
