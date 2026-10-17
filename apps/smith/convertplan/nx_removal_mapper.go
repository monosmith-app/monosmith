package convertplan

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// NxPackage reports whether a package belongs to Nx, so a converted
// workspace no longer needs it.
func NxPackage(name string) bool {
	for _, prefix := range []string{"@nx/", "@nrwl/", "@nx-go/", "@monodon/", "@nxlv/", "@mnci/nx-"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	switch name {
	case "nx", "nx-cloud", "@mnci/cli":
		return true
	}
	return false
}

// JestPreset replaces a jest.preset.js that re-exported @nx/jest/preset:
// Nx's preset values. Its resolver is replaced by the one thing a
// workspace without Nx needs from it: an ESM-style "./x.js" import of a
// TypeScript file resolves to "./x" (npm workspaces resolve the packages).
const JestPreset = `// Written by smith convert: the values @nx/jest/preset provided, without Nx.
module.exports = {
  testMatch: ['**/?(*.)+(spec|test).?([mc])[jt]s?(x)'],
  moduleFileExtensions: ['ts', 'js', 'mts', 'mjs', 'cts', 'cjs', 'html'],
  coverageReporters: ['html'],
  moduleNameMapper: { '^(\\.{1,2}/.*)\\.js$': '$1' },
  transform: {
    '^.+\\.(ts|js|mts|mjs|cts|cjs|html)$': ['ts-jest', { tsconfig: '<rootDir>/tsconfig.spec.json' }],
  },
  testEnvironment: 'jsdom',
  modulePathIgnorePatterns: ['<rootDir>/dist/', '<rootDir>/out-tsc/'],
  testEnvironmentOptions: {
    customExportConditions: ['node', 'require', 'default'],
  },
}
`

// configFile reports whether a file is configuration or tooling that may
// import Nx at run time.
func configFile(f string) bool {
	base := path.Base(f)
	for _, skip := range []string{"node_modules", ".nx", "testdata", "fixtures", "__fixtures__"} {
		if strings.HasPrefix(f, skip+"/") || strings.Contains(f, "/"+skip+"/") {
			return false
		}
	}
	for _, ext := range []string{".js", ".cjs", ".mjs", ".ts", ".cts", ".mts"} {
		if !strings.HasSuffix(base, ext) {
			continue
		}
		stem := strings.TrimSuffix(base, ext)
		if strings.HasSuffix(stem, ".config") || strings.HasPrefix(stem, "jest.preset") || strings.HasPrefix(f, "tools/") {
			return true
		}
	}
	return strings.HasPrefix(base, ".eslintrc")
}

// planNxRemoval schedules every Nx file and dependency for removal, rewrites
// the configuration that only re-exported Nx, and reports the rest.
func planNxRemoval(plan *Plan, w *nxworkspace.Workspace, files []string, read func(string) string) {
	plan.DeletePath("nx.json")
	plan.DeletePath(".nx")
	for _, p := range w.Projects {
		if pp := plan.ProjectAt(p.Root); p.HasProjectJSON && pp != nil && pp.Supported {
			plan.DeletePath(p.Root + "/project.json")
		}
	}
	for _, f := range files {
		if f == "project.json" {
			plan.DeletePath(f)
		}
	}
	var remove []string
	for _, deps := range []map[string]string{w.RootPackage.DevDependencies, w.RootPackage.Dependencies} {
		for name := range deps {
			if NxPackage(name) {
				remove = append(remove, name)
			}
		}
	}
	sort.Strings(remove)
	plan.RemoveDependencies = remove
	for name := range w.RootPackage.Overrides {
		if NxPackage(name) {
			plan.RemoveOverrides = append(plan.RemoveOverrides, name)
		}
	}
	sort.Strings(plan.RemoveOverrides)
	if len(remove) > 0 {
		plan.Add(Mapped, "", "package.json", "removes "+strings.Join(remove, ", "))
	}

	for _, f := range files {
		switch f {
		case "pnpm-lock.yaml", "pnpm-workspace.yaml":
			plan.Add(Unsupported, "", f, "pnpm workspaces: the smith Node plugin runs npm; switch to npm workspaces (npm install creates package-lock.json)")
		case "yarn.lock":
			plan.Add(Unsupported, "", f, "yarn workspaces: the smith Node plugin runs npm; switch to npm workspaces (npm install creates package-lock.json)")
		}
		if tsconfigFile(f) {
			planNxReactTypings(plan, f, read(f), files)
			continue
		}
		if !configFile(f) {
			continue
		}
		text := read(f)
		if !strings.Contains(text, "@nx/") && !strings.Contains(text, "'nx/") && !strings.Contains(text, "\"nx/") && !strings.Contains(text, "@nrwl/") {
			continue
		}
		base := path.Base(f)
		switch {
		case path.Dir(f) == "." && strings.HasPrefix(base, "jest.preset.") && strings.Contains(text, "@nx/jest/preset"):
			plan.WriteFile(f, JestPreset)
			plan.Add(Changed, "", f, "rewritten without @nx/jest: the same preset values, written out")
		case path.Dir(f) == "." && strings.HasPrefix(base, "jest.config.") && strings.Contains(text, "getJestProjects"):
			plan.DeletePath(f)
			plan.Add(Dropped, "", f, "listed every Jest project for Nx; smith runs each project's tests itself")
		case strings.HasPrefix(base, "jest.config.") && strings.Contains(text, "@nx/react/"):
			rewritten, ok := RewriteNxReactJest(text, path.Dir(f))
			if !ok {
				plan.Add(Manual, "", f, "imports Nx ("+nxImport(text)+"), which convert removes; rewrite it without Nx")
				break
			}
			plan.WriteFile(f, rewritten)
			plan.WriteFile(AssetTransformFile, AssetTransform)
			if plan.AddDevDependencies == nil {
				plan.AddDevDependencies = map[string]string{}
			}
			for _, d := range []string{"babel-jest", "@babel/preset-env", "@babel/preset-typescript", "@babel/preset-react"} {
				plan.AddDevDependencies[d] = ""
			}
			plan.Add(Changed, "", f, "rewritten without @nx/react: its Jest transformer is "+AssetTransformFile+" at the workspace root, and @nx/react/babel becomes the Babel presets it stood for (without its decorator, macro and runtime plugins)")
		case strings.HasPrefix(base, "rollup.config.") && strings.Contains(text, "with-nx"):
			if why := replaceRollup(plan, w, f, text); why != "" {
				plan.Add(Manual, "", f, why)
			}
		case strings.HasPrefix(base, "eslint.config."):
			rewritten, ok := RewriteNxESLint(text)
			if !ok {
				plan.Add(Manual, "", f, "uses @nx/eslint-plugin in a way convert cannot rewrite; replace it with typescript-eslint")
				break
			}
			plan.WriteFile(f, rewritten)
			if strings.Contains(rewritten, "typescript-eslint") {
				if plan.AddDevDependencies == nil {
					plan.AddDevDependencies = map[string]string{}
				}
				plan.AddDevDependencies["@eslint/js"] = ""
				plan.AddDevDependencies["typescript-eslint"] = ""
			}
			if strings.Contains(rewritten, "typescript-eslint") {
				plan.Add(Changed, "", f, "rewritten without @nx/eslint-plugin: its presets become @eslint/js and typescript-eslint, and the module-boundary rule (it reads the Nx graph) is removed")
			} else {
				plan.Add(Changed, "", f, "the @nx/dependency-checks rule (it reads the Nx graph) is removed")
			}
		case pathDeleted(plan, f):
		default:
			plan.Add(Manual, "", f, "imports Nx ("+nxImport(text)+"), which convert removes; rewrite it without Nx")
		}
	}
	if w.Nx.NxCloudID != "" || w.Nx.NxCloudToken != "" {
		plan.Add(Changed, "", "nx.json", "Nx Cloud is not used: smith caches locally (.smith/cache)")
	}
}

var (
	rollupMain     = regexp.MustCompile(`main:\s*['"]([^'"]+)['"]`)
	rollupFormat   = regexp.MustCompile(`format:\s*\[\s*['"](\w+)['"]`)
	rollupTSConfig = regexp.MustCompile(`tsConfig:\s*['"]([^'"]+)['"]`)
)

// replaceRollup swaps a library's @nx/rollup build for esbuild (the bundle)
// plus tsc (the declarations), writing to the paths its package.json
// already publishes. It returns why it could not, or "".
func replaceRollup(plan *Plan, w *nxworkspace.Workspace, file, text string) string {
	const unreadable = "builds with @nx/rollup's withNx, which needs Nx; rewrite the build without it"
	root := path.Dir(file)
	var project *nxworkspace.Project
	for _, p := range w.Projects {
		if p.Root == root {
			project = p
		}
	}
	pp := plan.ProjectAt(root)
	if project == nil || project.Package == nil || pp == nil || pp.Language != "node" {
		return unreadable
	}
	main := rollupMain.FindStringSubmatch(text)
	tsconfig := rollupTSConfig.FindStringSubmatch(text)
	out := strings.TrimPrefix(project.Package.Module, "./")
	if out == "" {
		out = strings.TrimPrefix(project.Package.Main, "./")
	}
	types := strings.TrimPrefix(project.Package.Types, "./")
	if main == nil || tsconfig == nil || out == "" || types == "" {
		return unreadable
	}
	format := "esm"
	if m := rollupFormat.FindStringSubmatch(text); m != nil {
		format = m[1]
	}
	entry := strings.TrimPrefix(main[1], "./")
	bundle := fmt.Sprintf("esbuild %s --bundle --packages=external --format=%s --outfile=%s --sourcemap", entry, format, out)
	entries, readable := RollupEntries(text)
	if !readable {
		return "builds several entries (input or additionalEntryPoints) convert cannot read; build each with esbuild, e.g. esbuild name=src/file.ts ... --splitting --outdir=" + path.Dir(out)
	}
	if len(entries) > 0 {
		multi, ok := rollupMultiEntry(entries, format, out)
		if !ok {
			return fmt.Sprintf("builds %d entries, but %s (package.json main or module) is none of them as %s output; build them with esbuild by hand", len(entries), out, format)
		}
		bundle = multi
	}
	// The declarations keep their own build info: sharing the project's
	// would make `tsc --build` (typecheck) think the library is built when
	// its own outputs are missing.
	declarations := path.Dir(types)
	up := strings.Repeat("../", strings.Count(path.Clean(root), "/")+1)
	if root == "." {
		up = ""
	}
	// Then each relative specifier in the declarations names its file, as
	// mnci's rollup plugin did, so nodenext consumers can follow them.
	build := fmt.Sprintf("%s && tsc -p %s --emitDeclarationOnly --declaration --declarationDir %s --tsBuildInfoFile %s/tsconfig.lib.tsbuildinfo && node %s%s %s",
		bundle, strings.TrimPrefix(tsconfig[1], "./"), declarations, declarations, up, DeclarationSpecifiersFile, declarations)
	plan.WriteFile(DeclarationSpecifiersFile, DeclarationSpecifiers)
	if pp.Scripts == nil {
		pp.Scripts = map[string]string{}
	}
	pp.Scripts["build"] = build
	if plan.AddDevDependencies == nil {
		plan.AddDevDependencies = map[string]string{}
	}
	plan.AddDevDependencies["esbuild"] = ""
	plan.DeletePath(file)
	built := out
	if len(entries) > 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name
		}
		built = fmt.Sprintf("every entry Rollup built (%s, sharing code in %s/chunks)", strings.Join(names, ", "), path.Dir(out))
	}
	plan.Add(Changed, pp.Name, "build", "Rollup with @nx/rollup becomes esbuild plus tsc declarations, into the same "+built+" and "+types+", whose relative specifiers "+DeclarationSpecifiersFile+" gives their .js files (as the old mnci's rollup plugin did): "+build)
	return ""
}

// rollupMultiEntry builds every entry of a rollup config in one esbuild run.
// With --splitting the code two entries share sits in one chunk, as Rollup
// did, so a class two subpaths export keeps one identity. The output names
// follow out (package.json main or module): its entry's name plus the
// extension Rollup gave it (index.esm.js: .esm.js). esbuild splits ES
// modules only, so another format reports false.
func rollupMultiEntry(entries []RollupEntry, format, out string) (string, bool) {
	if format != "esm" {
		return "", false
	}
	base := path.Base(out)
	ext := ""
	for _, e := range entries {
		if strings.HasPrefix(base, e.Name+".") && strings.HasSuffix(base, ".js") {
			ext = strings.TrimPrefix(base, e.Name)
		}
	}
	if ext == "" {
		return "", false
	}
	parts := []string{"esbuild"}
	for _, e := range entries {
		parts = append(parts, e.Name+"="+e.Source)
	}
	parts = append(parts, "--bundle", "--splitting", "--packages=external", "--format=esm", "--outdir="+path.Dir(out))
	if ext != ".js" {
		parts = append(parts, "--out-extension:.js="+ext)
	}
	parts = append(parts, "--chunk-names=chunks/[name]-[hash]", "--sourcemap")
	return strings.Join(parts, " "), true
}

// tsconfigFile reports whether a file is a tsconfig outside node_modules.
func tsconfigFile(f string) bool {
	base := path.Base(f)
	return strings.HasPrefix(base, "tsconfig") && strings.HasSuffix(base, ".json") &&
		!strings.HasPrefix(f, "node_modules/") && !strings.Contains(f, "/node_modules/")
}

// planNxReactTypings rewrites a tsconfig that lists @nx/react's typings,
// which go with @nx/react. A Vite project gets vite/client, which declares
// the same CSS-module and image imports.
func planNxReactTypings(plan *Plan, f, text string, files []string) {
	if !strings.Contains(text, nxReactTypings) {
		return
	}
	dir := path.Dir(f)
	vite := false
	for _, other := range files {
		if path.Dir(other) == dir && strings.HasPrefix(path.Base(other), "vite.config.") {
			vite = true
		}
	}
	rewritten, ok := RewriteNxReactTypings(text)
	if !vite || !ok {
		plan.Add(Manual, "", f, "lists @nx/react's typings ("+nxReactTypings+"*.d.ts), which go with @nx/react; declare the CSS-module and image imports in the project instead")
		return
	}
	plan.WriteFile(f, rewritten)
	plan.Add(Changed, "", f, "@nx/react's typings become vite/client, which declares the same CSS-module and image imports; an SVG import is its URL only (no ReactComponent export)")
}

func pathDeleted(plan *Plan, f string) bool {
	for _, d := range plan.Delete {
		if d == f || strings.HasPrefix(f, d+"/") {
			return true
		}
	}
	return false
}

// nxImport names the first Nx module a file imports.
func nxImport(text string) string {
	for _, marker := range []string{"@nx/", "@nrwl/", "'nx/", "\"nx/"} {
		i := strings.Index(text, marker)
		if i < 0 {
			continue
		}
		start := i
		if marker[0] == '\'' || marker[0] == '"' {
			start++
		}
		end := strings.IndexAny(text[start:], "'\"` \n")
		if end < 0 {
			end = len(text) - start
		}
		return text[start : start+end]
	}
	return "nx"
}
