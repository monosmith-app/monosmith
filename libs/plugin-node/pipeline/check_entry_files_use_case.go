package pipeline

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckEntryFiles checks that a package's build produced every file its
// package.json main, module, types and exports name in its build output
// (the project's dist/, or the workspace's dist/<project dir>): a missing
// one breaks only for whoever installs the package. Files outside the
// output, such as a source an exports condition points at, are not the
// build's. It judges a package whose build has run, so some of the files
// exist; with none of them, the package is not built yet.
func CheckEntryFiles(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if _, builds := p.Tasks["build"]; !builds || p.Kind != protocol.ProjectPackage {
		return protocol.Finding{}, false
	}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "package.json"))
	if err != nil {
		return protocol.Finding{}, false
	}
	var pkg struct {
		Main, Module, Types, Typings string
		Exports                      json.RawMessage
		Scripts                      map[string]string
	}
	if json.Unmarshal(raw, &pkg) != nil {
		return protocol.Finding{}, false
	}
	files := map[string]bool{}
	for _, f := range []string{pkg.Main, pkg.Module, pkg.Types, pkg.Typings} {
		addEntryFile(files, p.Root, f)
	}
	var exports any
	if len(pkg.Exports) > 0 && json.Unmarshal(pkg.Exports, &exports) == nil {
		collectExportFiles(files, p.Root, exports)
	}
	var missing []string
	found := false
	for f := range files {
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(f))); err == nil {
			found = true
		} else {
			missing = append(missing, f)
		}
	}
	if !found {
		return protocol.Finding{}, false
	}
	sort.Strings(missing)
	finding := protocol.Finding{
		Check:  p.Name + " builds every file its package.json names",
		OK:     len(missing) == 0,
		Detail: p.Root + "/package.json names " + strings.Join(missing, ", ") + ", which the build did not produce",
		Remedy: "build each entry the exports name (several entries in one esbuild run: esbuild name=src/file.ts ... --splitting --outdir=dist), or remove the entries from package.json; rebuild from clean to check (an old build's files can hide a missing one)",
	}
	var declarations []string
	for _, f := range missing {
		if isDeclarationFile(f) {
			declarations = append(declarations, f)
		}
	}
	if len(declarations) > 0 {
		finding.Remedy = declarationRemedy(workspaceRoot, p, pkg.Scripts["build"], declarations, pkg.Types, &finding)
		if len(declarations) < len(missing) {
			finding.Remedy += "\nand build each other entry the exports name, or remove it from package.json"
		}
	}
	return finding, true
}

// declarationEmit finds a build script that already writes type declarations:
// tsc, or a bundler's dts option.
var declarationEmit = regexp.MustCompile(`\btsc\b|\btsup\b|--dts\b|\bdts\b`)

// isDeclarationFile reports a TypeScript declaration file (.d.ts, .d.mts,
// .d.cts).
func isDeclarationFile(f string) bool {
	for _, ext := range []string{".d.ts", ".d.mts", ".d.cts"} {
		if strings.HasSuffix(f, ext) {
			return true
		}
	}
	return false
}

// declarationRemedy says how the build is to write the declarations the
// package.json names. A bundler (esbuild) writes none, so tsc is to; when the
// build script already runs tsc or a dts option, its output is somewhere
// else than the package.json says. A build that only lacks tsc gets the fix
// that appends it.
func declarationRemedy(workspaceRoot string, p protocol.Project, build string, missing []string, types string, finding *protocol.Finding) string {
	if declarationEmit.MatchString(build) {
		return "the build script already writes declarations, but not to the files package.json names: make tsc's --declarationDir (or rootDir) put them there, or point types and exports at where they are; rebuild from clean to check"
	}
	dir := ""
	if types != "" {
		dir = path.Dir(path.Clean(types))
	} else if rel, err := filepath.Rel(filepath.FromSlash(p.Root), filepath.FromSlash(missing[0])); err == nil {
		dir = path.Dir(filepath.ToSlash(rel))
	}
	if dir != "" && dir != "." {
		finding.Fix = manifestfixes.EmitDeclarationsFix(workspaceRoot, p, dir)
	}
	return "emit the declarations in the build: a bundler such as esbuild writes none, so append `&& tsc -p tsconfig.build.json --emitDeclarationOnly --declaration --declarationDir <the directory of the types entry>` to the build script (smith doctor --fix does); or remove the types entries from package.json"
}

// collectExportFiles walks package.json exports: a path, a list of
// fallbacks, or conditions and subpaths mapping to either.
func collectExportFiles(files map[string]bool, root string, v any) {
	switch e := v.(type) {
	case string:
		addEntryFile(files, root, e)
	case []any:
		for _, item := range e {
			collectExportFiles(files, root, item)
		}
	case map[string]any:
		for _, item := range e {
			collectExportFiles(files, root, item)
		}
	}
}

// addEntryFile keeps, workspace-relative, a file the build makes: one in
// the build output, and not a pattern.
func addEntryFile(files map[string]bool, root, f string) {
	if f == "" || strings.Contains(f, "*") {
		return
	}
	resolved := path.Clean(path.Join(root, f))
	if strings.HasPrefix(resolved, path.Clean(root)+"/dist/") || protocol.InBuildDir(resolved) {
		files[resolved] = true
	}
}
