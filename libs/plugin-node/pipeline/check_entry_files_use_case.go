package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckEntryFiles checks that a package's build produced every file its
// package.json main, module, types and exports name: a missing one breaks
// only for whoever installs the package. It judges a package whose build
// has run, so some of the files exist; with none of them, the package is
// not built yet and there is nothing to check.
func CheckEntryFiles(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if _, builds := p.Tasks["build"]; !builds || p.Kind != protocol.ProjectPackage {
		return protocol.Finding{}, false
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return protocol.Finding{}, false
	}
	var pkg struct {
		Main, Module, Types, Typings string
		Exports                      json.RawMessage
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return protocol.Finding{}, false
	}
	files := map[string]bool{}
	for _, f := range []string{pkg.Main, pkg.Module, pkg.Types, pkg.Typings} {
		addEntryFile(files, f)
	}
	var exports any
	if len(pkg.Exports) > 0 && json.Unmarshal(pkg.Exports, &exports) == nil {
		collectExportFiles(files, exports)
	}
	var missing []string
	found := false
	for f := range files {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(f))); err == nil {
			found = true
		} else {
			missing = append(missing, f)
		}
	}
	if !found {
		return protocol.Finding{}, false
	}
	sort.Strings(missing)
	return protocol.Finding{
		Check:  p.Name + " builds every file its package.json names",
		OK:     len(missing) == 0,
		Detail: p.Root + "/package.json names " + strings.Join(missing, ", ") + ", which the build did not produce",
		Remedy: "build each entry the exports name (several entries in one esbuild run: esbuild name=src/file.ts ... --splitting --outdir=dist), or remove the entries from package.json; rebuild from clean to check (an old build's files can hide a missing one)",
	}, true
}

// collectExportFiles walks package.json exports: a path, a list of
// fallbacks, or conditions and subpaths mapping to either.
func collectExportFiles(files map[string]bool, v any) {
	switch e := v.(type) {
	case string:
		addEntryFile(files, e)
	case []any:
		for _, item := range e {
			collectExportFiles(files, item)
		}
	case map[string]any:
		for _, item := range e {
			collectExportFiles(files, item)
		}
	}
}

// addEntryFile keeps a file the build makes: not a pattern, and not the
// package.json the package ships as is.
func addEntryFile(files map[string]bool, f string) {
	f = path.Clean(strings.TrimPrefix(f, "./"))
	if f == "." || f == "package.json" || strings.Contains(f, "*") || strings.HasPrefix(f, "../") {
		return
	}
	files[f] = true
}
