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
	ownDist := buildsIntoWorkspaceDist(workspaceRoot, p)
	for _, f := range []string{pkg.Main, pkg.Module, pkg.Types, pkg.Typings} {
		addEntryFile(files, p.Root, f, ownDist)
	}
	var exports any
	if len(pkg.Exports) > 0 && json.Unmarshal(pkg.Exports, &exports) == nil {
		collectExportFiles(files, p.Root, exports, ownDist)
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
		Check:   p.Name + " builds every file its package.json names",
		Project: p.Name,
		OK:      len(missing) == 0,
		Detail:  p.Root + "/package.json names " + strings.Join(missing, ", ") + ", which the build did not produce",
		Remedy:  "build each entry the exports name (several entries in one esbuild run: esbuild name=src/file.ts ... --splitting --outdir=dist), or remove the entries from package.json; rebuild from clean to check (an old build's files can hide a missing one)",
	}
	if !finding.OK {
		finding.Problem = p.Root + "/package.json names " + strings.Join(missing, ", ") + ", which the build does not produce"
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
		return "the build script already writes declarations, but not to " + missing[0] + ": " + declarationDiagnosis(workspaceRoot, p, build, missing[0]) + "; rebuild from clean to check"
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

var (
	// tscProject is the tsconfig a tsc command of the build script reads.
	tscProject = regexp.MustCompile(`\btsc\b[^&|;]*?\s(?:-p|--project)[= ]+"?([^"\s&;|]+)"?`)
	// declarationOn finds what makes tsc write declarations, in a tsconfig
	// or on the command line.
	declarationOn = regexp.MustCompile(`"(?:declaration|composite)"\s*:\s*true|--declaration\b|--emitDeclarationOnly\b|\s-d\b`)
	rootDirOption = regexp.MustCompile(`"rootDir"\s*:\s*"([^"]*)"`)
	includeOption = regexp.MustCompile(`(?s)"include"\s*:\s*(\[[^\]]*\])`)
)

// declarationDiagnosis says why tsc did not write want, a declaration file
// the package.json names, by reading the tsconfig the build gives tsc and
// looking for the file under another path of the build output.
func declarationDiagnosis(workspaceRoot string, p protocol.Project, build, want string) string {
	if !regexp.MustCompile(`\btsc\b`).MatchString(build) {
		return "make the tool that writes them put them there, or point types and exports at where they are"
	}
	name := "tsconfig.json"
	if m := tscProject.FindStringSubmatch(build); m != nil {
		name = m[1]
	}
	config := path.Join(p.Root, filepath.ToSlash(name))
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(config)))
	if err != nil {
		return "the build runs tsc with " + config + ", which does not exist: create it with the package's sources in its include, or point -p at the tsconfig that has them"
	}
	text := string(raw)
	if !declarationOn.MatchString(text) && !declarationOn.MatchString(build) {
		return config + " does not set \"declaration\": true, so tsc writes no declarations: add it under compilerOptions"
	}
	outRoot := p.Root + "/dist"
	if strings.HasPrefix(want, protocol.ProjectBuildDir(p.Root)+"/") {
		outRoot = protocol.ProjectBuildDir(p.Root)
	}
	if found := findByName(workspaceRoot, outRoot, path.Base(want)); len(found) > 0 {
		rootDir := "not set, so tsc keeps the folders from the common folder of the files it compiles"
		if m := rootDirOption.FindStringSubmatch(text); m != nil {
			rootDir = `"` + m[1] + `"`
		}
		return "tsc wrote " + strings.Join(found, ", ") + " instead. Point types and the types of exports in package.json at ./dist/" +
			strings.TrimPrefix(found[0], outRoot+"/") + ", or change rootDir in " + config + " (now " + rootDir + ") so the path matches"
	}
	include := "not set, so tsc compiles every .ts file beside it"
	if m := includeOption.FindStringSubmatch(text); m != nil {
		include = strings.Join(strings.Fields(m[1]), " ")
	}
	return "tsc wrote no " + path.Base(want) + " under " + outRoot + ": the include of " + config + " (now " + include + ") must hold the source of that entry, such as src/index.ts"
}

// findByName lists, workspace-relative and sorted, the files named base under
// dir (at most three).
func findByName(workspaceRoot, dir, base string) []string {
	var found []string
	_ = filepath.WalkDir(filepath.Join(workspaceRoot, filepath.FromSlash(dir)), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == base && len(found) < 3 {
			if rel, err := filepath.Rel(workspaceRoot, p); err == nil {
				found = append(found, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	sort.Strings(found)
	return found
}

// collectExportFiles walks package.json exports: a path, a list of
// fallbacks, or conditions and subpaths mapping to either.
func collectExportFiles(files map[string]bool, root string, v any, ownDist bool) {
	switch e := v.(type) {
	case string:
		addEntryFile(files, root, e, ownDist)
	case []any:
		for _, item := range e {
			collectExportFiles(files, root, item, ownDist)
		}
	case map[string]any:
		for _, item := range e {
			collectExportFiles(files, root, item, ownDist)
		}
	}
}

// buildsIntoWorkspaceDist reports whether the project's build writes into
// dist/<project dir>, so that the ./dist/... of its package.json (the form it
// is published in) names files of that folder.
func buildsIntoWorkspaceDist(workspaceRoot string, p protocol.Project) bool {
	_, outDir, ok := manifestfixes.LocateOutDir(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
	return ok && protocol.InBuildDir(path.Clean(path.Join(p.Root, filepath.ToSlash(outDir))))
}

// addEntryFile keeps, workspace-relative, a file the build makes: one in
// the build output, and not a pattern. With ownDist, ./dist/x of the package
// is dist/<project dir>/x of the workspace.
func addEntryFile(files map[string]bool, root, f string, ownDist bool) {
	if f == "" || strings.Contains(f, "*") {
		return
	}
	resolved := path.Clean(path.Join(root, f))
	if own := path.Clean(root) + "/dist/"; ownDist && strings.HasPrefix(resolved, own) {
		resolved = protocol.ProjectBuildDir(root) + "/" + strings.TrimPrefix(resolved, own)
	}
	if strings.HasPrefix(resolved, path.Clean(root)+"/dist/") || protocol.InBuildDir(resolved) {
		files[resolved] = true
	}
}
