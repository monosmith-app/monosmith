package releasepack

import (
	"path"
	"strings"
)

// Ecosystem describes how one language's package files are recognised and
// installed from a folder.
type Ecosystem struct {
	// Name is how the ecosystem is shown: npm, Python, NuGet, RubyGems, crates.
	Name string
	// Patterns are the gh --pattern globs that select its files.
	Patterns []string
	// Extensions are the file name endings of its package files.
	Extensions []string
}

// ecosystems maps a language plugin to its ecosystem.
var ecosystems = map[string]Ecosystem{
	"node":   {Name: "npm", Patterns: []string{"*.tgz"}, Extensions: []string{".tgz"}},
	"python": {Name: "Python", Patterns: []string{"*.whl", "*.tar.gz"}, Extensions: []string{".whl", ".tar.gz"}},
	"dotnet": {Name: "NuGet", Patterns: []string{"*.nupkg"}, Extensions: []string{".nupkg"}},
	"ruby":   {Name: "RubyGems", Patterns: []string{"*.gem"}, Extensions: []string{".gem"}},
	"rust":   {Name: "crates", Patterns: []string{"*.crate"}, Extensions: []string{".crate"}},
}

// EcosystemOf finds the ecosystem of a language plugin; false when the plugin
// has no package file to attach (Go, PHP, JVM and Dart release by their tag).
func EcosystemOf(plugin string) (Ecosystem, bool) {
	e, ok := ecosystems[plugin]
	return e, ok
}

// IsPackageFile reports whether an asset name is a package file of the
// ecosystem (not its checksum file).
func (e Ecosystem) IsPackageFile(name string) bool {
	for _, ext := range e.Extensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

// installCommand is the command that installs the package files of an
// ecosystem from the folder they were downloaded to.
func installCommand(e Ecosystem, files []string) string {
	switch e.Name {
	case "npm":
		return "npm i " + joinLocal(files)
	case "Python":
		wheels := filter(files, ".whl")
		if len(wheels) == 0 {
			wheels = files
		}
		return "pip install " + joinLocal(wheels)
	case "RubyGems":
		return "gem install " + joinLocal(files)
	case "NuGet":
		var lines []string
		for _, f := range files {
			id, version := nugetIDAndVersion(f)
			lines = append(lines, "dotnet add package "+id+" --version "+version+" --source .")
		}
		return strings.Join(lines, "\n")
	case "crates":
		var lines []string
		for _, f := range files {
			lines = append(lines, "tar -xzf ./"+f)
		}
		return strings.Join(lines, "\n") + "\n# then depend on each folder by path, for example: " + crateDependency(files[0])
	}
	return ""
}

// crateDependency is the Cargo.toml line that depends on an unpacked crate.
func crateDependency(file string) string {
	folder := strings.TrimSuffix(path.Base(file), ".crate")
	name := folder
	if i := strings.LastIndex(folder, "-"); i > 0 {
		name = folder[:i]
	}
	return name + ` = { path = "` + folder + `" }`
}

// nugetIDAndVersion splits Acme.Core.1.0.0.nupkg into its id and version.
func nugetIDAndVersion(file string) (string, string) {
	base := strings.TrimSuffix(path.Base(file), ".nupkg")
	parts := strings.Split(base, ".")
	for i, part := range parts {
		if i > 0 && part != "" && part[0] >= '0' && part[0] <= '9' {
			return strings.Join(parts[:i], "."), strings.Join(parts[i:], ".")
		}
	}
	return base, ""
}

func joinLocal(files []string) string {
	local := make([]string, len(files))
	for i, f := range files {
		local[i] = "./" + f
	}
	return strings.Join(local, " ")
}

func filter(files []string, suffix string) []string {
	var out []string
	for _, f := range files {
		if strings.HasSuffix(f, suffix) {
			out = append(out, f)
		}
	}
	return out
}
