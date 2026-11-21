package outsidefiles

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Ref is a path a manifest names, and the kind of thing it names.
type Ref struct {
	// Value is the path as written.
	Value string
	// Kind is KindFile or KindProject.
	Kind string
}

var (
	// includeAttr matches an MSBuild element that takes a path in Include or
	// Project: <None Include="..\LICENSE">, <Import Project="..\x.props" />.
	includeAttr = regexp.MustCompile(`<([A-Za-z][\w.]*)\b[^>]*?\b(?:Include|Project)\s*=\s*"([^"]+)"`)
	// pathProperty matches an MSBuild property whose value is a file path.
	pathProperty = regexp.MustCompile(`<(AssemblyOriginatorKeyFile|ApplicationIcon|ApplicationManifest|CodeAnalysisRuleSet|Win32Resource)>\s*([^<]+?)\s*</`)
)

// MSBuildRefs returns the paths an MSBuild file (a .csproj, .fsproj, .props
// or .targets) names that leave its folder: item and Import paths that start
// with .., and path properties such as AssemblyOriginatorKeyFile. A path with
// a property, an item or a wildcard ($(, @(, %(, *, ?) is skipped, because only
// MSBuild knows its value.
//
// Parameters: text is the content of the file.
// Returns: the references in the order of the file, each once.
func MSBuildRefs(text string) []Ref {
	var out []Ref
	seen := map[string]bool{}
	add := func(value, kind string) {
		value = strings.TrimSpace(value)
		if seen[value] || !leaves(value) || strings.ContainsAny(value, "$@%*?") {
			return
		}
		seen[value] = true
		out = append(out, Ref{Value: value, Kind: kind})
	}
	for _, m := range includeAttr.FindAllStringSubmatch(text, -1) {
		kind := KindFile
		if m[1] == "ProjectReference" {
			kind = KindProject
		}
		for _, part := range strings.Split(m[2], ";") {
			add(part, kind)
		}
	}
	for _, m := range pathProperty.FindAllStringSubmatch(text, -1) {
		add(m[2], KindFile)
	}
	return out
}

// PackageJSONRefs returns the "files" entries of a package.json that leave its
// folder (../LICENSE).
//
// Parameters: raw is the content of the package.json.
// Returns: the references, or nil for a file that is not JSON or has no such
// entry.
func PackageJSONRefs(raw []byte) []Ref {
	var pkg struct {
		Files []string `json:"files"`
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return nil
	}
	var out []Ref
	for _, f := range pkg.Files {
		if leaves(f) && !strings.ContainsAny(f, "*?") {
			out = append(out, Ref{Value: f, Kind: KindFile})
		}
	}
	return out
}

// leaves reports whether a relative path starts by going up a folder.
func leaves(value string) bool {
	return strings.HasPrefix(strings.ReplaceAll(value, `\`, "/"), "../")
}
