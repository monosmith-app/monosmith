package projecttooling

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
)

// Assessment is what Assess found for one project.
type Assessment struct {
	// Name is the project's package name.
	Name string
	// Dir is the project folder, workspace-relative with forward slashes ("."
	// for a project at the root).
	Dir string
	// Node is false when the folder has no package.json: the step does not
	// apply to the project.
	Node bool
	// Kind is app, lib or package.
	Kind string
	// Facts are what its files say.
	Facts Facts
	// Legacy lists the legacy tools in words (LegacyReasons).
	Legacy []string
	// Recommended is the answer smith recommends.
	Recommended Answer
}

// Assess reads one project and says which of its build and lint tools are
// legacy, and so which answer smith recommends.
//
// Parameters: root is the workspace folder (absolute); dir is the project
// folder, workspace-relative ("." for a project at the root).
// Returns: the assessment (Node false when dir has no package.json); an error
// for a package.json that is not JSON.
func Assess(root, dir string) (Assessment, error) {
	dir = cleanRel(dir)
	a := Assessment{Dir: dir}
	abs := filepath.Join(root, filepath.FromSlash(dir))
	raw, err := os.ReadFile(filepath.Join(abs, "package.json"))
	if err != nil {
		return a, nil
	}
	var pkg discovery.PackageManifest
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return a, err
	}
	a.Node, a.Name = true, pkg.Name
	if a.Name == "" {
		a.Name = path.Base(filepath.ToSlash(abs))
	}
	f := Facts{BuildScript: pkg.Scripts["build"], Dependencies: map[string]bool{}}
	f.ESLintInManifest = len(pkg.EslintConfig) > 0 && string(pkg.EslintConfig) != "null"
	for _, deps := range []map[string]string{pkg.Dependencies, pkg.DevDependencies} {
		for name, r := range deps {
			f.Dependencies[name] = true
			if name == "eslint" {
				f.ESLintRange = r
			}
		}
	}
	if f.ESLintRange == "" && dir != "." {
		var rootPkg discovery.PackageManifest
		if raw, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil && json.Unmarshal(raw, &rootPkg) == nil {
			f.ESLintRange = rootPkg.DevDependencies["eslint"]
		}
	}
	f.LegacyESLintFiles = legacyESLintFiles(root, dir)
	if _, out, ok := discovery.LocateOutDir(abs); ok {
		f.OutDir = out
	}
	a.Facts = f
	a.Kind = kindOf(dir, pkg)
	a.Legacy = LegacyReasons(f)
	a.Recommended = Recommend(a.Legacy)
	return a, nil
}

// legacyESLintFiles lists the .eslintrc files (a kept .eslintrc.js.pre-smith
// too) of the project folder and of the workspace root.
func legacyESLintFiles(root, dir string) []string {
	var out []string
	for _, d := range []string{dir, "."} {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(d)))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), ".eslintrc") {
				out = append(out, path.Join(d, e.Name()))
			}
		}
		if d == "." {
			break
		}
	}
	sort.Strings(out)
	return out
}

// kindOf is the project's kind: its smith.kind, else its folder, else a
// create-react-app project is an app, a private package a lib, and a public
// one a package.
func kindOf(dir string, pkg discovery.PackageManifest) string {
	if pkg.Smith.Kind != "" {
		return pkg.Smith.Kind
	}
	for _, k := range []string{"app", "lib", "package"} {
		if strings.HasPrefix(dir, k+"s/") {
			return k
		}
	}
	if _, ok := pkg.Dependencies["react-scripts"]; ok {
		return "app"
	}
	if pkg.Private {
		return "lib"
	}
	return "package"
}

// cleanRel is a workspace-relative folder with forward slashes, "." for the root.
func cleanRel(dir string) string {
	if dir == "" {
		return "."
	}
	return path.Clean(filepath.ToSlash(dir))
}
