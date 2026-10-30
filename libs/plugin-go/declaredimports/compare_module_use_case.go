package declaredimports

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-go/dependencies"
	"github.com/monosmith-app/monosmith/libs/plugin-go/pipeline"
)

// Compare checks go.mod of the module at moduleRoot (workspace-relative)
// against the imports of its sources, without running the go command and
// without writing anything. It is what `go mod tidy` would change, as far as
// the imports tell:
//
//   - an imported package no requirement provides is Missing;
//   - a requirement not marked indirect that nothing imports is Unused;
//   - a requirement marked indirect that the module imports is MarkedIndirect.
//
// Packages of the module itself and of the other modules go.work uses are
// never missing: the workspace provides them (tidy cannot see go.work, so a
// module that leans on it is checked here and not by tidy). Tool directives
// and `go run <package>` in a go:generate line count as imports; versions are
// not compared.
//
// The second result says the module imports or requires a go.work member.
func Compare(workspaceRoot, moduleRoot string) (Result, bool, error) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(moduleRoot))
	raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return Result{}, false, err
	}
	gomod := string(raw)
	own := modulePathOf(gomod)
	requires := dependencies.ParseRequirements(gomod)
	members := workspaceModules(workspaceRoot, moduleRoot)
	imports, err := scanImports(dir)
	if err != nil {
		return Result{}, false, err
	}
	for _, tool := range toolPackages(gomod) {
		imports = append(imports, sourceImport{File: "go.mod", Path: tool})
	}

	result := Result{Method: "go.mod compared with the imports"}
	used := map[string]bool{}
	reported := map[string]bool{}
	leans := false
	for _, imp := range imports {
		path := imp.Path
		switch {
		case path == "C" || path == "unsafe" || !strings.Contains(firstElement(path), "."):
			continue // cgo and the standard library
		case within(path, own):
			continue
		}
		if member := providerOf(path, members); member != "" {
			leans = true
			used[member] = true
			continue
		}
		req, ok := providerIn(path, requires)
		switch {
		case !ok:
			if !reported[path] {
				reported[path] = true
				result.Issues = append(result.Issues, Issue{Kind: Missing, Path: path, File: imp.File})
			}
		case req.Indirect:
			used[req.Path] = true
			if !reported[req.Path] {
				reported[req.Path] = true
				result.Issues = append(result.Issues, Issue{Kind: MarkedIndirect, Path: req.Path, File: imp.File})
			}
		default:
			used[req.Path] = true
		}
	}
	for _, req := range requires {
		if providerOf(req.Path, members) != "" {
			leans = true
		}
		if !req.Indirect && !used[req.Path] {
			result.Issues = append(result.Issues, Issue{Kind: Unused, Path: req.Path})
		}
	}
	sort.SliceStable(result.Issues, func(i, j int) bool {
		a, b := result.Issues[i], result.Issues[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Path < b.Path
	})
	return result, leans, nil
}

// workspaceModules are the module paths of the other modules go.work uses,
// none when the workspace has no go.work.
func workspaceModules(workspaceRoot, moduleRoot string) []string {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "go.work"))
	if err != nil {
		return nil
	}
	var out []string
	for _, entry := range pipeline.ParseGoWork(string(src)).Use {
		if filepath.ToSlash(filepath.Clean(entry)) == moduleRoot {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(entry), "go.mod"))
		if err != nil {
			continue
		}
		if path := modulePathOf(string(raw)); path != "" {
			out = append(out, path)
		}
	}
	return out
}

func firstElement(path string) string {
	first, _, _ := strings.Cut(path, "/")
	return first
}

// within reports whether the package path is the module or inside it.
func within(path, module string) bool {
	return module != "" && (path == module || strings.HasPrefix(path, module+"/"))
}

// providerOf returns the module among modules that contains the package, or
// "" (the longest, as a nested module's path extends its parent's).
func providerOf(path string, modules []string) string {
	best := ""
	for _, m := range modules {
		if within(path, m) && len(m) > len(best) {
			best = m
		}
	}
	return best
}

func providerIn(path string, requires []dependencies.Requirement) (dependencies.Requirement, bool) {
	var best dependencies.Requirement
	found := false
	for _, r := range requires {
		if within(path, r.Path) && (!found || len(r.Path) > len(best.Path)) {
			best, found = r, true
		}
	}
	return best, found
}
