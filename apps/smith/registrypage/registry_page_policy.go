package registrypage

import (
	"net/url"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Page is the address of a package's page on its registry, or "" when it cannot
// be known: a link that may be wrong is worse than none.
//
// The registry follows the language plugin. A project that installs from a
// private registry named in smith.toml ([registries.<name>]) uses that one: its
// web page is derived for Azure Artifacts and is none for any other private
// registry. A package that comes from the public registry has its page there.
//
// Parameters: plugin is the language plugin that serves the project (node,
// python, dotnet, rust, jvm, dart, ruby, php, go); name is the package as the
// plugin names it (groupId:artifactId for jvm); install is the registry the
// project installs from, nil for the public one; ambientPrivate is true when
// the workspace points the tool at a private registry some other way (an
// .npmrc, PIP_INDEX_URL, a NuGet.Config), which AmbientPrivate tells.
// Returns: the https address, or "".
func Page(plugin, name string, install *protocol.Registry, ambientPrivate bool) string {
	if name == "" {
		return ""
	}
	if install != nil {
		if raw := installURL(plugin, install); raw != "" {
			return privatePage(plugin, name, raw)
		}
	}
	if ambientPrivate {
		return ""
	}
	return publicPage(plugin, name)
}

// installURL is the registry URL the project sets for the plugin's ecosystem;
// "" when smith.toml sets none for it (the public registry is used).
func installURL(plugin string, r *protocol.Registry) string {
	switch plugin {
	case "node":
		return r.NPM
	case "python":
		return r.PyPI
	case "dotnet":
		return r.NuGet
	}
	return ""
}

// publicPage is the page on the registry of the language.
func publicPage(plugin, name string) string {
	switch plugin {
	case "node":
		return "https://www.npmjs.com/package/" + name
	case "python":
		return "https://pypi.org/project/" + url.PathEscape(name) + "/"
	case "dotnet":
		return "https://www.nuget.org/packages/" + url.PathEscape(name)
	case "rust":
		return "https://crates.io/crates/" + url.PathEscape(name)
	case "jvm":
		group, artifact, ok := strings.Cut(name, ":")
		if !ok || group == "" || artifact == "" {
			return ""
		}
		return "https://central.sonatype.com/artifact/" + url.PathEscape(group) + "/" + url.PathEscape(artifact)
	case "dart":
		return "https://pub.dev/packages/" + url.PathEscape(name)
	case "ruby":
		return "https://rubygems.org/gems/" + url.PathEscape(name)
	case "php":
		return "https://packagist.org/packages/" + name
	case "go":
		return "https://pkg.go.dev/" + name
	}
	return ""
}

// publicHosts are the hosts of the public registries of the three ecosystems
// that can be pointed elsewhere; a project that names one of them has the
// public page.
var publicHosts = map[string]bool{
	"registry.npmjs.org": true, "registry.yarnpkg.com": true,
	"pypi.org": true, "files.pythonhosted.org": true,
	"api.nuget.org": true,
}

// azureSegment is how Azure Artifacts names the ecosystem in a package address.
var azureSegment = map[string]string{"node": "Npm", "python": "PyPI", "dotnet": "NuGet"}

// privatePage derives the page of a package on a registry the project names.
func privatePage(plugin, name, raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	if publicHosts[host] {
		return publicPage(plugin, name)
	}
	if host == "pkgs.dev.azure.com" {
		return azurePage(plugin, name, u)
	}
	return ""
}

// azurePage is the page on Azure Artifacts. The feed address is
// https://pkgs.dev.azure.com/<org>[/<project>]/_packaging/<feed>[@<view>]/<kind>/...;
// the page is https://dev.azure.com/<org>[/<project>]/_artifacts/feed/<feed>/<Kind>/<name>/overview.
func azurePage(plugin, name string, u *url.URL) string {
	segment, ok := azureSegment[plugin]
	if !ok {
		return ""
	}
	parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
	at := -1
	for i, p := range parts {
		if p == "_packaging" {
			at = i
		}
	}
	if at < 1 || at > 2 || at+1 >= len(parts) {
		return ""
	}
	feed, _, _ := strings.Cut(parts[at+1], "@")
	if feed == "" {
		return ""
	}
	return "https://dev.azure.com/" + strings.Join(parts[:at], "/") + "/_artifacts/feed/" + feed + "/" + segment + "/" + url.PathEscape(name) + "/overview"
}
