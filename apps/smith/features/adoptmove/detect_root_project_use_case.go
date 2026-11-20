package adoptmove

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	modulePath  = regexp.MustCompile(`(?m)^module\s+("?)([^\s"]+)`)
	tomlName    = regexp.MustCompile(`(?m)^name\s*=\s*["']([^"']+)["']`)
	yamlName    = regexp.MustCompile(`(?m)^name:\s*["']?([^\s"']+)`)
	packageMain = regexp.MustCompile(`(?m)^package\s+main\b`)
	dotnetApp   = regexp.MustCompile(`(?i)<OutputType>\s*(Exe|WinExe)\s*</OutputType>|Sdk="Microsoft\.NET\.Sdk\.(Web|Worker)"`)
)

// Detect looks at the top level of the repository at dir and returns the one
// project it holds. A file there is a project's manifest when a language
// plugin claims its name (claims come from the plugins the session started, so
// adopt knows every language the workspace knows). It returns ErrNoRootProject
// when no claimed manifest is there, and a problem when the root holds the
// manifests of more than one language, because that is two projects in one
// folder. A root that is a workspace already (a package.json with
// "workspaces", a go.work, a Cargo [workspace]) is not one project:
// ErrNoRootProject. Nothing is changed.
//
// Parameters: dir is the repository root; claims are the language plugins'
// manifest names.
// Returns: the candidate, ErrNoRootProject, or a problem for an ambiguous root.
func Detect(dir string, claims []Claim) (Candidate, error) {
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(dir, rel))
		return err == nil
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(dir, rel))
		return string(raw)
	}
	entries, _ := os.ReadDir(dir)
	var found []string
	var candidate Candidate
	for _, c := range claims {
		for _, e := range entries {
			if e.IsDir() || !protocol.IsManifest(c.Manifests, e.Name()) {
				continue
			}
			found = append(found, e.Name())
			candidate = Candidate{Ecosystem: c.Language, Manifest: e.Name()}
			break
		}
	}
	switch {
	case len(found) == 0:
		return Candidate{}, ErrNoRootProject
	case len(found) > 1:
		return Candidate{}, problem.New("the root holds more than one project ("+strings.Join(found, ", ")+"), so smith cannot tell what to move",
			"split them by hand into a folder each (git mv), commit, then run smith adopt again",
			"or run smith adopt --init --move "+Leave+" for an empty workspace and add projects with smith add")
	}
	base := filepath.Base(dir)
	if abs, err := filepath.Abs(dir); err == nil {
		base = filepath.Base(abs)
	}
	text := read(candidate.Manifest)
	switch candidate.Ecosystem {
	case EcosystemNode:
		var pkg struct {
			Name       string
			Private    bool
			Workspaces json.RawMessage
			Bin        json.RawMessage
			Scripts    map[string]string
		}
		if json.Unmarshal([]byte(text), &pkg) != nil {
			return Candidate{}, problem.New("package.json at the root cannot be read as JSON", "fix package.json, then run smith adopt again")
		}
		if len(pkg.Workspaces) > 0 && string(pkg.Workspaces) != "null" && string(pkg.Workspaces) != "[]" {
			return Candidate{}, ErrNoRootProject
		}
		candidate.Name = DeriveName(pkg.Name, base)
		candidate.App = pkg.Scripts["start"] != "" || (len(pkg.Bin) > 0 && string(pkg.Bin) != "null")
		candidate.Published = pkg.Name != "" && !pkg.Private
	case EcosystemGo:
		if has("go.work") {
			return Candidate{}, ErrNoRootProject
		}
		if m := modulePath.FindStringSubmatch(text); m != nil {
			candidate.Name = DeriveName(m[2], base)
		} else {
			candidate.Name = DeriveName("", base)
		}
		candidate.App = rootHasMainPackage(dir)
	case EcosystemPython:
		candidate.Name = DeriveName(firstTomlName(text), base)
		candidate.App = strings.Contains(text, "[project.scripts]") || strings.Contains(text, "[tool.poetry.scripts]") || has("__main__.py") || has("main.py")
	case EcosystemRust:
		if strings.Contains(text, "[workspace]") {
			return Candidate{}, ErrNoRootProject
		}
		candidate.Name = DeriveName(firstTomlName(text), base)
		candidate.App = has(filepath.Join("src", "main.rs")) || strings.Contains(text, "[[bin]]")
	default:
		candidate.Name = DeriveName(manifestName(candidate.Manifest, text), base)
		candidate.App = manifestIsApp(candidate.Manifest, text)
	}
	return candidate, nil
}

// manifestName is the project name a manifest of another language gives: the
// "name" of a JSON, TOML or YAML manifest, else the file's stem for a manifest
// named after its project (App.csproj, gem.gemspec). It is "" when the
// manifest names nothing, and the folder name is used.
func manifestName(file, text string) string {
	ext := path.Ext(file)
	switch ext {
	case ".json":
		var m struct{ Name string }
		_ = json.Unmarshal([]byte(text), &m)
		return m.Name
	case ".toml":
		return firstTomlName(text)
	case ".yaml", ".yml":
		if m := yamlName.FindStringSubmatch(text); m != nil {
			return m[1]
		}
		return ""
	case ".csproj", ".fsproj", ".gemspec", ".rockspec", ".podspec":
		return strings.TrimSuffix(file, ext)
	}
	return ""
}

// manifestIsApp says whether a manifest of another language describes a
// program: a .NET project whose output is an executable or that uses the web
// or worker SDK, or a Composer "project". Any other manifest is offered as a
// package, and the person may choose the app folder.
func manifestIsApp(file, text string) bool {
	switch {
	case strings.HasSuffix(file, ".csproj") || strings.HasSuffix(file, ".fsproj"):
		return dotnetApp.MatchString(text)
	case file == "composer.json":
		var m struct{ Type string }
		_ = json.Unmarshal([]byte(text), &m)
		return m.Type == "project"
	}
	return false
}

// firstTomlName is the first name = "..." of a manifest, which is the
// [project] or [package] table's in the files adopt reads.
func firstTomlName(text string) string {
	if m := tomlName.FindStringSubmatch(text); m != nil {
		return m[1]
	}
	return ""
}

// rootHasMainPackage is true when a .go file at the top of dir declares package main.
func rootHasMainPackage(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err == nil && packageMain.Match(raw) {
			return true
		}
	}
	return false
}
