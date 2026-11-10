package manifestfixes

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// LibraryBuildDescription names the grouped fix, so that doctor --fix --only
// can pick it.
const LibraryBuildDescription = "build each library into the workspace's dist folder and resolve it by name to its sources"

// LibraryPlan is what doctor offers for the library packages that build into
// their own folder: one fix for the workspace, or the reason a package is left.
type LibraryPlan struct {
	// Fix is the one grouped fix; nil when no package can move.
	Fix *protocol.Fix
	// Packages are the names of the projects the fix moves.
	Packages []string
	// Refusals gives, for a package the plan covers but cannot move, the reason.
	Refusals map[string]string
}

// Moves reports whether the fix moves the project.
func (l LibraryPlan) Moves(name string) bool {
	for _, p := range l.Packages {
		if p == name {
			return true
		}
	}
	return false
}

// libraryMove is the edits that move one package.
type libraryMove struct {
	project protocol.Project
	old     string
	alias   string // the package name
	source  string // its entry, relative to the workspace
	files   []protocol.FileEdit
}

var (
	// tomlGroupBuildOutputs is a group table that only gives the build task the old folder as outputs.
	tomlGroupBuildOutputs = regexp.MustCompile(`(?m)^\[groups\.tasks\.build\][ \t]*\r?\noutputs[ \t]*=[ \t]*\["dist"\][ \t]*\r?\n(\r?\n|\[|\z)`)
	sourceFile            = regexp.MustCompile(`\.[cm]?tsx?$`)
)

// PlanLibraryMoves plans the move of every library package whose package.json
// leads into its own build folder and that other packages import by name (or
// that has "exports", which cannot name a folder outside the package): the
// Nx layout. The build goes to dist/<project dir>; package.json keeps its
// main, types and exports as they are, the form that is published (Stage
// makes them relative to the folder it packs); and the root tsconfig.base.json
// maps each package name to its source entry, so tsc, Jest and bundlers in the
// repository read the sources.
//
// A package smith cannot rewrite with certainty is not moved and gets a
// reason in Refusals: a build script it cannot read, a second output folder,
// exports subpaths, an alias already mapped elsewhere. The plan is idempotent:
// a package that builds into dist/<project dir> is not a candidate.
//
// Parameters: workspaceRoot is the absolute workspace folder; projects the
// Node projects; dependents the projects that import each project, by name.
// Returns the plan; its Fix is a question, never a silent change: the owner
// picks "move" or "leave", and the files are shown line by line first.
func PlanLibraryMoves(workspaceRoot string, projects []protocol.Project, dependents map[string][]string) LibraryPlan {
	plan := LibraryPlan{Refusals: map[string]string{}}
	var moves []libraryMove
	for _, p := range projects {
		move, refusal, candidate := planLibrary(workspaceRoot, p, len(dependents[p.Name]) > 0)
		switch {
		case !candidate:
		case refusal != "":
			plan.Refusals[p.Name] = refusal
		default:
			moves = append(moves, *move)
		}
	}
	if len(moves) == 0 {
		return plan
	}
	baseText, err := os.ReadFile(filepath.Join(workspaceRoot, "tsconfig.base.json"))
	if err != nil {
		return refuseAll(plan, moves, "tsconfig.base.json is missing at the workspace root, so there is no place to map the package names to their sources")
	}
	aliases := map[string]string{}
	for _, m := range moves {
		// tsc wants a path with no baseUrl to start with "./".
		aliases[m.alias] = "./" + m.source
	}
	_, conflicts, err := addPathAliases(string(baseText), aliases)
	if err != nil {
		return refuseAll(plan, moves, "tsconfig.base.json cannot be edited: "+err.Error())
	}
	kept := moves[:0]
	for _, m := range moves {
		if have, clash := conflicts[m.alias]; clash {
			plan.Refusals[m.project.Name] = "tsconfig.base.json maps " + m.alias + " to " + have + " already, not to " + m.source + "; make them agree by hand"
			delete(aliases, m.alias)
			continue
		}
		kept = append(kept, m)
	}
	moves = kept
	if len(moves) == 0 {
		return plan
	}
	edited, _, _ := addPathAliases(string(baseText), aliases)
	files := []protocol.FileEdit{}
	var steps []protocol.FixStep
	var needsGroup []protocol.Project
	for _, m := range moves {
		files = append(files, m.files...)
		plan.Packages = append(plan.Packages, m.project.Name)
		if step, ok := dropDeclaredOutputs(workspaceRoot, m.project, m.old); ok {
			steps = append(steps, step)
		} else if len(m.project.Tasks["build"].Outputs) > 0 {
			needsGroup = append(needsGroup, m.project)
		}
	}
	if edited != string(baseText) {
		files = append(files, protocol.FileEdit{Path: "tsconfig.base.json", Content: edited})
	}
	notes := []string{}
	if len(needsGroup) > 0 {
		toml, err := os.ReadFile(filepath.Join(workspaceRoot, "smith.toml"))
		switch {
		case err != nil || len(plan.Refusals) > 0 || !tomlGroupBuildOutputs.Match(toml):
			notes = append(notes, "smith.toml gives the build task of "+needsGroup[0].Name+" and others the outputs of the old folder, from a group; smith does not edit that: remove the outputs line by hand, or the build caches a folder nothing writes")
		default:
			files = append(files, protocol.FileEdit{Path: "smith.toml", Content: dropGroupBuildOutputs(string(toml))})
			notes = append(notes, "the outputs = [\"dist\"] of [groups.tasks.build] in smith.toml goes, so the build caches dist/<project dir> again")
		}
	}
	jestQuestion, jestNote := jestChoice(workspaceRoot, moves)
	if jestNote != "" {
		notes = append(notes, jestNote)
	}
	move := protocol.Fix{Description: LibraryBuildDescription, Files: files, Steps: steps}
	if jestQuestion != nil {
		move.Choices = []protocol.FixChoice{*jestQuestion}
	}
	for _, r := range sortedKeys(plan.Refusals) {
		notes = append(notes, r+" stays as it is: "+plan.Refusals[r])
	}
	names := strings.Join(plan.Packages, ", ")
	plan.Fix = &protocol.Fix{
		Description: LibraryBuildDescription,
		AppliesTo:   plan.Packages,
		Choices: []protocol.FixChoice{{
			Question: "These library packages build into their own folder, inside the source tree: " + names + ". smith builds every project into dist/<project dir>. Moving them edits, for each, the build script in package.json and the outDir, declarationDir and tsBuildInfoFile of its tsconfig files to ../../dist/<project dir>. Its main, types and exports stay: they are the form that is published, and smith packs from dist/<project dir> with them made relative to it. tsconfig.base.json gets one paths entry per package name that points at its source entry, so importers in the repository read the source, not a build. " + strings.Join(notes, ". ") + ". The next screen shows every changed line. Nothing is deleted or committed: delete the old <project dir>/dist folders yourself, because Jest resolves a leftover one before the sources.",
			Options: []protocol.FixChoiceOption{
				{Label: "move the build of " + countOf(len(plan.Packages), "package") + " into dist/ and resolve them by name to their sources", Fix: move},
				{Label: "leave them as they are (nothing changes)"},
			},
		}},
	}
	return plan
}

// refuseAll gives every move the same reason.
func refuseAll(plan LibraryPlan, moves []libraryMove, reason string) LibraryPlan {
	for _, m := range moves {
		plan.Refusals[m.project.Name] = reason
	}
	return plan
}

func countOf(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// dropGroupBuildOutputs removes the [groups.tasks.build] table that only sets
// outputs = ["dist"].
func dropGroupBuildOutputs(toml string) string {
	return tomlGroupBuildOutputs.ReplaceAllStringFunc(toml, func(m string) string {
		tail := tomlGroupBuildOutputs.FindStringSubmatch(m)[1]
		if strings.TrimSpace(tail) == "" {
			return "" // the blank line that ended the table goes with it
		}
		return tail
	})
}

// planLibrary plans one package. candidate is false when the project is not a
// library this plan covers (not a package, builds in dist/<project dir>
// already, or reaches nothing in its build folder from package.json); the
// refusal says why a candidate cannot move.
func planLibrary(workspaceRoot string, p protocol.Project, imported bool) (move *libraryMove, refusal string, candidate bool) {
	if p.Kind != protocol.ProjectPackage {
		return nil, "", false
	}
	if _, builds := p.Tasks["build"]; !builds {
		return nil, "", false
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	_, outDir, ok := LocateOutDir(dir)
	if !ok {
		return nil, "", false
	}
	old := cleanDir(outDir)
	if old == "." || old == "" || strings.HasPrefix(old, "..") || protocol.InBuildDir(path.Clean(path.Join(p.Root, old))) {
		return nil, "", false
	}
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, "", false
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil, "", false
	}
	var manifest struct {
		Name    string            `json:"name"`
		Types   string            `json:"types"`
		Typings string            `json:"typings"`
		Exports json.RawMessage   `json:"exports"`
		Scripts map[string]string `json:"scripts"`
	}
	_ = json.Unmarshal(raw, &manifest)
	into := false
	reason := ""
	doc.RewriteStrings(buildPathKeys, func(v string) string {
		switch {
		case under(v, old):
			into = true
		case v != "./package.json" && reason == "":
			reason = "package.json points at " + v + ", which is not in " + old + ", the build folder"
		}
		return v
	})
	if !into || (!imported && len(manifest.Exports) == 0) {
		return nil, "", false
	}
	if reason != "" {
		return nil, reason, true
	}
	if manifest.Name == "" {
		return nil, "package.json has no name to map to the source", true
	}
	if r := exportsRefusal(manifest.Exports); r != "" {
		return nil, r, true
	}
	newDir := protocol.BuildDirFromProject(p.Root)
	build := manifest.Scripts["build"]
	moved, why := moveLibraryScript(build, old, newDir)
	if why != "" {
		return nil, "its build script is not one smith rewrites: " + why, true
	}
	if moved == build {
		return nil, "its build script names no folder in " + old, true
	}
	for name, script := range manifest.Scripts {
		if name != "build" && namesFolder(script, old) {
			return nil, "its " + name + " script names " + old + ", and smith does not rewrite it; edit it by hand", true
		}
	}
	entry := libraryEntry(dir, build, manifest.Types, manifest.Typings, manifest.Exports, old)
	if entry == "" {
		return nil, "smith cannot find the source file of the package's main entry (an esbuild entry or the types file's source in src/)", true
	}
	doc.Object("scripts", false).SetString("build", moved)
	files := []protocol.FileEdit{{Path: path.Join(p.Root, "package.json"), Content: render(doc, raw)}}
	configs, _ := filepath.Glob(filepath.Join(dir, "tsconfig*.json"))
	declared := declarationFlag(build)
	for _, file := range configs {
		text, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		edited := moveConfigOutputs(string(text), old, newDir)
		if declared != "" && usesConfig(build, filepath.Base(file)) && !configHasOption(edited, "declarationDir") {
			edited = addDeclarationDir(edited, followPath(declared, old, newDir))
		}
		if edited != string(text) {
			files = append(files, protocol.FileEdit{Path: path.Join(p.Root, filepath.Base(file)), Content: edited})
		}
	}
	return &libraryMove{project: p, old: old, alias: manifest.Name, source: path.Join(p.Root, entry), files: files}, "", true
}

// exportsRefusal is why the exports field cannot be mapped to one source
// entry, or "".
func exportsRefusal(raw json.RawMessage) string {
	var exports map[string]any
	if len(raw) == 0 || json.Unmarshal(raw, &exports) != nil {
		return ""
	}
	for key := range exports {
		if strings.HasPrefix(key, ".") && key != "." && key != "./package.json" {
			return "package.json exports has the subpath " + key + ", and smith maps only the main entry to a source"
		}
	}
	return ""
}

// libraryEntry is the source file of the package's main entry, relative to the
// project: the file the build script gives esbuild, or the source that the
// types file in the build folder comes from.
func libraryEntry(dir, build, types, typings string, exports json.RawMessage, old string) string {
	for _, part := range strings.Split(build, " && ") {
		if programOf(part) != "esbuild" {
			continue
		}
		for _, token := range strings.Fields(part)[1:] {
			if token = unquote(token); !strings.HasPrefix(token, "-") && sourceFile.MatchString(token) && fileExists(filepath.Join(dir, filepath.FromSlash(token))) {
				return cleanDir(token)
			}
		}
	}
	declarations := []string{types, typings}
	var tree any
	if json.Unmarshal(exports, &tree) == nil {
		if m, ok := tree.(map[string]any); ok {
			if dot, ok := m["."].(map[string]any); ok {
				tree = dot
			}
			if conditions, ok := tree.(map[string]any); ok {
				if s, ok := conditions["types"].(string); ok {
					declarations = append([]string{s}, declarations...)
				}
			}
		}
	}
	for _, d := range declarations {
		if !under(d, old) {
			continue
		}
		rel := strings.TrimPrefix(cleanDir(d), old+"/")
		for _, suffix := range []string{".d.ts", ".d.mts", ".d.cts"} {
			if !strings.HasSuffix(rel, suffix) {
				continue
			}
			base := strings.TrimSuffix(rel, suffix)
			for _, candidate := range []string{base + ".ts", "src/" + base + ".ts", strings.TrimPrefix(base, "src/") + ".ts"} {
				if fileExists(filepath.Join(dir, filepath.FromSlash(candidate))) {
					return candidate
				}
			}
		}
	}
	return ""
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// declarationFlag is the folder the build script gives tsc as --declarationDir.
func declarationFlag(build string) string {
	if m := regexp.MustCompile(`--declarationDir[= ]+"?([^"\s&;|]+)"?`).FindStringSubmatch(build); m != nil {
		return cleanDir(m[1])
	}
	return ""
}

// usesConfig reports whether a tsc command of the script reads the tsconfig.
func usesConfig(build, name string) bool {
	for _, part := range strings.Split(build, " && ") {
		if programOf(part) == "tsc" {
			if m := projectFlag.FindStringSubmatch(part); m != nil && path.Base(cleanDir(m[1])) == name {
				return true
			}
		}
	}
	return false
}

// jestChoice reads the Jest preset of the workspace. When it does not read
// tsconfig paths, it returns the question that offers a moduleNameMapper; the
// note says what smith found otherwise.
func jestChoice(workspaceRoot string, moves []libraryMove) (*protocol.FixChoice, string) {
	for _, name := range []string{"jest.preset.js", "jest.preset.cjs"} {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, name))
		if err != nil {
			continue
		}
		text := string(raw)
		switch {
		case strings.Contains(text, "@monosmith/jest-preset"):
			return nil, name + " uses @monosmith/jest-preset, whose resolver reads the tsconfig paths, so Jest needs no edit"
		case !strings.Contains(text, "module.exports"):
			return nil, name + " is not a CommonJS file smith can append to, so map the package names to their sources in Jest by hand (moduleNameMapper)"
		}
		eol := "\n"
		if strings.Contains(text, "\r\n") {
			eol = "\r\n"
		}
		lines := []string{
			"",
			"// Written by smith doctor: a workspace package resolves to its source in tests, as in tsconfig.base.json paths.",
			"module.exports.moduleNameMapper = {",
			"  ...module.exports.moduleNameMapper,",
		}
		sort.Slice(moves, func(i, j int) bool { return moves[i].alias < moves[j].alias })
		for _, m := range moves {
			key := "^" + regexp.QuoteMeta(m.alias) + "$"
			lines = append(lines, "  '"+strings.ReplaceAll(key, `\`, `\\`)+"': require('node:path').join(__dirname, '"+m.source+"'),")
		}
		lines = append(lines, "}", "")
		appended := strings.TrimRight(text, "\r\n") + eol + strings.Join(lines, eol)
		return &protocol.FixChoice{
			Question: name + " does not read the tsconfig paths, so Jest would find each package through node_modules and its package.json, which lead into a build folder that no longer exists. How should Jest find the packages?",
			Options: []protocol.FixChoiceOption{
				{Label: "add a moduleNameMapper entry per package to " + name + ", pointing at its source", Fix: protocol.Fix{Files: []protocol.FileEdit{{Path: name, Content: appended}}}},
				{Label: "leave Jest as it is (tests that import a package by name fail until you map it)"},
			},
		}, ""
	}
	return nil, ""
}
