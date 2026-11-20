package projecttooling

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/vite"
	"github.com/monosmith-app/monosmith/libs/plugin-node/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/libs/plugin-node/toolchainversions"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/eslintpeer"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// jsdomRange is the DOM that Vitest runs a React app's tests in.
const jsdomRange = "^26.1.0"

// PlaceSmithBuild puts smith's build in place for the project: a tsc build
// into lib/ (or another folder) builds into dist (a lib or package inside its
// own folder, an app into the workspace's dist/, ADR 0102), with main, types
// and files following; a create-react-app project builds with Vite (and its
// tests run with Vitest). A build smith does not rewrite (rollup, webpack, a
// script of its own) stays, and the outcome says so. It removes no package
// and no config: the removal plan does that once this step is Ready.
//
// Parameters: root is the workspace folder (absolute); a is the project's
// assessment; out gets one line for each change.
// Returns: the outcome; an error for a file that cannot be read or written.
func PlaceSmithBuild(root string, a Assessment, out io.Writer) (Outcome, error) {
	var placed Outcome
	var err error
	switch ClassifyBuild(a.Facts) {
	case BuildNone, BuildSmith:
		return Outcome{Ready: true}, nil
	case BuildTsc:
		placed, err = placeTscBuild(root, a, out)
	case BuildCRA:
		placed, err = placeVite(root, a, out)
	default:
		return otherBuild(a), nil
	}
	if err != nil || !placed.Ready {
		return placed, err
	}
	return placed, ignoreFolders(root, out, "dist", "node_modules")
}

// ignoreFolders adds each folder name to the root .gitignore when no line
// ignores it at every depth: the build now writes into dist, and a project
// in a subfolder installs into its own node_modules, which a "/node_modules"
// rule anchored at the root does not hide.
func ignoreFolders(root string, out io.Writer, names ...string) error {
	file := filepath.Join(root, ".gitignore")
	raw, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	text := string(raw)
	var added []string
	for _, name := range names {
		if ignoredEverywhere(text, name) {
			continue
		}
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += name + "\n"
		added = append(added, name)
	}
	if len(added) == 0 {
		return nil
	}
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(out, "  build: .gitignore now ignores %s\n", strings.Join(added, " and "))
	return nil
}

// ignoredEverywhere reports whether a .gitignore line ignores the folder name
// at every depth.
func ignoredEverywhere(gitignore, name string) bool {
	for _, line := range strings.Split(gitignore, "\n") {
		switch strings.TrimSpace(line) {
		case name, name + "/", "**/" + name, "**/" + name + "/":
			return true
		}
	}
	return false
}

// otherBuild is the outcome of a build smith does not rewrite.
func otherBuild(a Assessment) Outcome {
	tool := bundlerOf(a.Facts.BuildScript)
	if tool == "" {
		tool = "a script of its own"
	}
	return Outcome{Left: []string{fmt.Sprintf("%s builds with %s (%q), which smith does not rewrite. Replace the build script with tsc or esbuild; its old build tools stay until then", a.Name, tool, a.Facts.BuildScript)}}
}

// placeTscBuild moves a tsc build into dist with the doctor's own fixes.
func placeTscBuild(root string, a Assessment, out io.Writer) (Outcome, error) {
	abs := filepath.Join(root, filepath.FromSlash(a.Dir))
	config, outDir, ok := discovery.LocateOutDir(abs)
	if !ok {
		return Outcome{Ready: true}, nil
	}
	p := protocol.Project{Name: a.Name, Root: a.Dir}
	var fix *protocol.Fix
	if a.Kind == string(protocol.ProjectApp) && a.Dir != "." {
		fix = manifestfixes.BuildIntoDistFix(root, p, config, outDir)
	} else {
		fix = manifestfixes.BuildInPlaceFix(root, p, config, outDir)
	}
	if fix == nil {
		return Outcome{Left: []string{fmt.Sprintf("%s builds into %s, and smith could not point its build at dist. Set outDir to dist in %s", a.Name, outDir, config)}}, nil
	}
	if err := writeEdits(root, fix.Files); err != nil {
		return Outcome{}, err
	}
	old := strings.TrimSuffix(strings.TrimPrefix(path.Clean(outDir), "./"), "/")
	if err := editManifest(filepath.Join(abs, "package.json"), func(doc *jsondoc.Document) error {
		return renameInFiles(doc, old, "dist")
	}); err != nil {
		return Outcome{}, err
	}
	fmt.Fprintf(out, "  build: %s builds into dist instead of %s (main, types and files follow)\n", a.Name, old)
	var left []string
	for _, s := range fix.Steps {
		left = append(left, a.Name+": "+s.Description+" (smith did not run it: run smith doctor --fix)")
	}
	return Outcome{Ready: true, Left: left}, nil
}

// renameInFiles points the package's "files" entries at the new build folder.
func renameInFiles(doc *jsondoc.Document, old, now string) error {
	if !doc.Has("files") {
		return nil
	}
	var probe struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(doc.Bytes(), &probe); err != nil {
		return nil
	}
	files := probe.Files
	changed := false
	for i, f := range files {
		clean := strings.TrimSuffix(strings.TrimPrefix(f, "./"), "/")
		if clean == old || strings.HasPrefix(clean, old+"/") {
			files[i] = now + strings.TrimPrefix(clean, old)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return doc.SetJSON("files", files)
}

// placeVite moves a create-react-app project onto Vite.
func placeVite(root string, a Assessment, out io.Writer) (Outcome, error) {
	abs := filepath.Join(root, filepath.FromSlash(a.Dir))
	page, err := os.ReadFile(filepath.Join(abs, "public", "index.html"))
	if err != nil {
		return Outcome{Left: []string{a.Name + " runs react-scripts but has no public/index.html, so smith cannot move it onto Vite. Write index.html and vite.config.ts by hand"}}, nil
	}
	entry := ""
	for _, name := range []string{"index.tsx", "index.ts", "index.jsx", "index.js"} {
		if _, err := os.Stat(filepath.Join(abs, "src", name)); err == nil {
			entry = "/src/" + name
			break
		}
	}
	if entry == "" {
		return Outcome{Left: []string{a.Name + " runs react-scripts but has no src/index.tsx (or .ts, .jsx, .js), so smith cannot move it onto Vite"}}, nil
	}
	src := scanSources(filepath.Join(abs, "src"))
	s := ViteSettings{OutDir: "dist", JSXInJS: src.jsxInJS, ProcessEnv: src.processEnv, Tests: len(src.tests) > 0}
	if a.Dir != "." {
		s.OutDir = protocol.BuildDirFromProject(a.Dir)
	}
	for _, name := range []string{"setupTests.ts", "setupTests.js"} {
		if _, err := os.Stat(filepath.Join(abs, "src", name)); err == nil && s.Tests {
			s.SetupFile = "src/" + name
		}
	}
	var homepage string
	if err := editManifest(filepath.Join(abs, "package.json"), func(doc *jsondoc.Document) error {
		homepage, _ = doc.String("homepage")
		scripts := doc.Object("scripts", true)
		for _, name := range scripts.Keys() {
			script, _ := scripts.String(name)
			next, keep := ViteScript(script, s.Tests)
			if !keep {
				scripts.Delete(name)
				continue
			}
			scripts.SetString(name, next)
		}
		if !scripts.Has("start") {
			scripts.SetString("start", "vite")
		}
		raiseTypesNode(doc, out)
		return nil
	}); err != nil {
		return Outcome{}, err
	}
	if u, err := url.Parse(homepage); err == nil && homepage != "" && u.Path != "" && u.Path != "/" {
		s.Base = strings.TrimSuffix(u.Path, "/") + "/"
	}
	edits := []protocol.FileEdit{
		{Path: path.Join(a.Dir, "index.html"), Content: ViteIndexHTML(string(page), entry)},
		{Path: path.Join(a.Dir, "public", "index.html"), Delete: true},
		{Path: path.Join(a.Dir, "vite.config.ts"), Content: ViteConfig(s)},
	}
	envFile := filepath.Join(abs, "src", "react-app-env.d.ts")
	if raw, err := os.ReadFile(envFile); err == nil && strings.Contains(string(raw), "react-scripts") {
		edits = append(edits, protocol.FileEdit{Path: path.Join(a.Dir, "src", "react-app-env.d.ts"), Content: "/// <reference types=\"vite/client\" />\n"})
	}
	if err := writeEdits(root, edits); err != nil {
		return Outcome{}, err
	}
	packages, overrides := vite.ReactBuildPackages()
	if s.Tests {
		packages["vitest"] = toolchainversions.Vitest
		packages["jsdom"] = jsdomRange
	}
	if err := editManifest(filepath.Join(root, "package.json"), func(doc *jsondoc.Document) error {
		addMissing(doc.Object("devDependencies", true), packages)
		if a.Dir != "." {
			raiseTypesNode(doc, out)
		}
		if len(overrides) > 0 {
			addMissing(doc.Object("overrides", true), overrides)
		}
		return nil
	}); err != nil {
		return Outcome{}, err
	}
	fmt.Fprintf(out, "  build: %s builds with Vite (index.html, vite.config.ts) into %s\n", a.Name, s.OutDir)
	var left []string
	switch {
	case s.Tests:
		fmt.Fprintf(out, "  test: %s runs its tests with Vitest (jsdom, Jest's globals)\n", a.Name)
	default:
		fmt.Fprintf(out, "  test: %s has no tests, so its react-scripts test script is gone (smith reports \"no tests\" as a warning)\n", a.Name)
	}
	if len(src.jestCalls) > 0 {
		left = append(left, fmt.Sprintf("%s's tests call jest.* (%s), which Vitest names vi.*: change jest.fn and the others to vi.fn", a.Name, strings.Join(src.jestCalls, ", ")))
	}
	if len(src.svgComponents) > 0 {
		left = append(left, fmt.Sprintf("%s imports SVG files as React components (%s), which Vite does without create-react-app only with vite-plugin-svgr: add it to vite.config.ts", a.Name, strings.Join(src.svgComponents, ", ")))
	}
	return Outcome{Ready: true, Left: left}, nil
}

// typesNodeFloor is the oldest @types/node that Vite accepts as its peer
// (^20.19.0 || >=22.12.0); an older one stops npm install with ERESOLVE.
const typesNodeFloor = 20

// raiseTypesNode sets an @types/node older than Vite accepts to the range
// smith's stacks pin, in the dependency sections of a package.json.
func raiseTypesNode(doc *jsondoc.Document, out io.Writer) {
	for _, section := range []string{"dependencies", "devDependencies"} {
		deps := doc.Object(section, false)
		if deps == nil {
			continue
		}
		current, ok := deps.String("@types/node")
		if !ok {
			continue
		}
		floor, simple := eslintpeer.RangeFloor(current)
		if !simple {
			continue
		}
		major, _, _ := strings.Cut(floor, ".")
		if n, err := strconv.Atoi(major); err != nil || n >= typesNodeFloor {
			continue
		}
		deps.SetString("@types/node", toolchainversions.TypesNode)
		fmt.Fprintf(out, "  build: @types/node %s → %s, which Vite needs\n", current, toolchainversions.TypesNode)
	}
}

// addMissing sets each name the object does not have yet, in name order.
func addMissing(obj *jsondoc.Document, values map[string]string) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !obj.Has(name) {
			obj.SetString(name, values[name])
		}
	}
}

// sources is what the source files of a create-react-app project use.
type sources struct {
	jsxInJS, processEnv bool
	tests               []string
	jestCalls           []string
	svgComponents       []string
}

// scanSources reads the files under src.
func scanSources(src string) sources {
	var s sources
	_ = filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".js" && ext != ".jsx" && ext != ".ts" && ext != ".tsx" {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		text := string(raw)
		rel, _ := filepath.Rel(filepath.Dir(src), p)
		rel = filepath.ToSlash(rel)
		base := filepath.Base(p)
		test := strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") || strings.Contains(rel, "__tests__/")
		if test {
			s.tests = append(s.tests, rel)
			if strings.Contains(text, "jest.") {
				s.jestCalls = append(s.jestCalls, rel)
			}
		}
		if ext == ".js" && (strings.Contains(text, "</") || strings.Contains(text, "/>")) {
			s.jsxInJS = true
		}
		if strings.Contains(text, "process.env") {
			s.processEnv = true
		}
		if strings.Contains(text, "ReactComponent") && strings.Contains(text, ".svg") {
			s.svgComponents = append(s.svgComponents, rel)
		}
		return nil
	})
	return s
}

// writeEdits writes or deletes files, workspace-relative.
func writeEdits(root string, edits []protocol.FileEdit) error {
	for _, e := range edits {
		full := filepath.Join(root, filepath.FromSlash(e.Path))
		if e.Delete {
			if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, []byte(e.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
