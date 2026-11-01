package adoptmove

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// rootManifests are the files that make the repository's root a project, by
// ecosystem. otherManifests are projects of languages smith has a plugin for
// that adopt cannot move yet; they count only to refuse an ambiguous root.
var (
	rootManifests = []struct{ file, ecosystem string }{
		{"package.json", EcosystemNode},
		{"go.mod", EcosystemGo},
		{"pyproject.toml", EcosystemPython},
		{"Cargo.toml", EcosystemRust},
	}
	otherManifests = []string{"pom.xml", "pubspec.yaml", "build.gradle", "build.gradle.kts"}

	modulePath  = regexp.MustCompile(`(?m)^module\s+("?)([^\s"]+)`)
	tomlName    = regexp.MustCompile(`(?m)^name\s*=\s*["']([^"']+)["']`)
	packageMain = regexp.MustCompile(`(?m)^package\s+main\b`)
)

// Detect looks at the top level of the repository at dir and returns the one
// project it holds. It returns ErrNoRootProject when no manifest is there, and
// a problem when the root is not one project smith can move: two manifests
// (a Go module with a package.json is two projects in one folder) or a manifest
// of a language adopt cannot move yet next to another. A root that is a
// workspace already (a package.json with "workspaces", a go.work, a Cargo
// [workspace]) is not one project: ErrNoRootProject. Nothing is changed.
func Detect(dir string) (Candidate, error) {
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(dir, rel))
		return err == nil
	}
	read := func(rel string) string {
		raw, _ := os.ReadFile(filepath.Join(dir, rel))
		return string(raw)
	}
	var found []string
	var candidate Candidate
	for _, m := range rootManifests {
		if has(m.file) {
			found = append(found, m.file)
			candidate = Candidate{Ecosystem: m.ecosystem, Manifest: m.file}
		}
	}
	for _, f := range otherManifests {
		if has(f) && len(found) > 0 {
			found = append(found, f)
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
	}
	return candidate, nil
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
