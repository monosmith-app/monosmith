package convertplan

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
	"github.com/monosmith-app/monosmith/apps/smith/release"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectkind"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// otherEcosystems name the manifests of languages smith v1 has no plugin for.
var otherEcosystems = []struct{ suffix, label string }{
	{"pubspec.yaml", "Flutter/Dart"}, {".csproj", ".NET"}, {".fsproj", ".NET"}, {"Cargo.toml", "Rust"},
	{"pom.xml", "Java (Maven)"}, {"build.gradle.kts", "JVM (Gradle)"}, {"build.gradle", "JVM (Gradle)"},
	{"composer.json", "PHP"}, {"Gemfile", "Ruby"}, {"CMakeLists.txt", "C++ (CMake)"}, {".bicep", "Bicep"},
	{Dockerfile, "a container image"}, {"requirements.txt", "Python without a pyproject.toml"},
}

// Dockerfile makes an app ship as an image, which the docker convention
// builds and releases.
const Dockerfile = "Dockerfile"

// Language decides which smith plugin owns a project, from the manifest at its
// root. label names the ecosystem when no plugin does.
func Language(p *nxworkspace.Project) (lang, label string) {
	switch {
	case p.HasFile("go.mod"):
		return "go", "Go"
	case p.HasFile("pyproject.toml"):
		return "python", "Python"
	}
	for _, e := range otherEcosystems {
		if e.suffix == Dockerfile && p.HasFile("package.json") {
			// A Node app shipped as an image: the docker convention builds it.
			continue
		}
		for _, f := range p.Files {
			if path.Dir(f) == p.Root && strings.HasSuffix(f, e.suffix) {
				return "", e.label
			}
		}
	}
	if p.HasFile("package.json") {
		return "node", "Node"
	}
	return "", "an Nx project with no manifest"
}

// Kind decides a project's smith kind: an application is an app, a project
// Nx releases is a package, anything else is an internal lib.
func Kind(p *nxworkspace.Project, released bool) protocol.ProjectKind {
	if p.Type == "app" || p.ProjectType == "application" {
		return protocol.ProjectApp
	}
	if released {
		return protocol.ProjectPackage
	}
	return protocol.ProjectLib
}

// NeedsKindMarker reports whether the directory alone would give the
// project another kind in smith, so the manifest must declare it.
func NeedsKindMarker(p *nxworkspace.Project, lang string, kind protocol.ProjectKind) bool {
	fallback := protocol.ProjectLib
	if lang == "node" && (p.Package == nil || !p.Package.Private) {
		fallback = protocol.ProjectPackage
	}
	return projectkind.Resolve("", p.Root, fallback) != kind
}

// TagPrefix is the release tag prefix that continues a project's existing
// tags: the Nx pattern's or smith's default, whichever has the newest release
// (a project renamed along the way keeps a stray old tag under the other
// name, which must not hide the newer ones). With no release under either,
// smith's default. An error means the pattern cannot become a prefix.
func TagPrefix(pattern string, p *nxworkspace.Project, smithDefault string, tags []string) (string, error) {
	prefix, ok := strings.CutSuffix(strings.ReplaceAll(pattern, "{projectName}", p.Name), "{version}")
	if !ok || strings.Contains(prefix, "{") {
		return smithDefault, fmt.Errorf("the release tag pattern %q cannot be continued (smith tags are <prefix><version>)", pattern)
	}
	if prefix == smithDefault {
		return smithDefault, nil
	}
	nx, hasNx := newestUnder(tags, prefix)
	def, hasDefault := newestUnder(tags, smithDefault)
	if hasNx && (!hasDefault || def.Less(nx)) {
		return prefix, nil
	}
	return smithDefault, nil
}

// newestUnder is the highest version among the tags that are prefix followed
// by a semantic version.
func newestUnder(tags []string, prefix string) (newest release.Version, found bool) {
	for _, t := range tags {
		rest, ok := strings.CutPrefix(t, prefix)
		if !ok {
			continue
		}
		v, err := release.ParseVersion(rest)
		if err != nil {
			continue
		}
		if !found || newest.Less(v) {
			newest, found = v, true
		}
	}
	return newest, found
}

// AppRelease reads the old mnci's release.asset of a releasable Go app.
func AppRelease(p *nxworkspace.Project, product string) *workspace.ProjectRelease {
	r := &workspace.ProjectRelease{AssetName: "go-app-{product}-{os}-{arch}.{ext}", Product: product}
	raw, ok := p.Release["asset"]
	if !ok {
		return r
	}
	var asset struct {
		Name      string            `json:"name"`
		Product   string            `json:"product"`
		OSAlias   map[string]string `json:"osAlias"`
		ArchAlias map[string]string `json:"archAlias"`
	}
	if json.Unmarshal(raw, &asset) != nil {
		return r
	}
	if asset.Name != "" {
		r.AssetName = asset.Name
	}
	if asset.Product != "" {
		r.Product = asset.Product
	}
	r.OSAlias, r.ArchAlias = asset.OSAlias, asset.ArchAlias
	return r
}
