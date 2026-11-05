package manifestfixes

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// MigrateBuildFix offers to move the build of a project that compiles outside
// the workspace's dist/ into dist/<project dir>. It is a question, never a
// silent change: the owner picks "move" (the files the move edits are shown
// line by line before anything is written) or "leave this project as it is".
//
// A package that is published (kind package, not private) is published from
// dist/<project dir> through Stage, which copies only that folder. A move that
// would leave a file of the tarball outside it is refused, with the reason,
// and no fix is returned.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project;
// configName the tsconfig that sets oldOutDir ("" when the build script does);
// oldOutDir the folder the build writes now, relative to the project.
// Returns the fix (nil when there is nothing to edit or the move is refused)
// and the refusal reason ("" unless refused).
func MigrateBuildFix(workspaceRoot string, p protocol.Project, configName, oldOutDir string, dependents []string) (*protocol.Fix, string) {
	if reason := MigrationRefusal(workspaceRoot, p, configName, oldOutDir, dependents); reason != "" {
		return nil, reason
	}
	move := BuildIntoDistFix(workspaceRoot, p, configName, oldOutDir)
	if move == nil {
		return nil, ""
	}
	move.AppliesTo = nil
	if step, ok := dropDeclaredOutputs(workspaceRoot, p, cleanDir(oldOutDir)); ok {
		move.Steps = append(move.Steps, step)
	}
	files := make([]string, len(move.Files))
	for i, f := range move.Files {
		files[i] = f.Path
	}
	target := protocol.ProjectBuildDir(p.Root)
	if len(move.Steps) > 0 {
		files = append(files, "and the outputs override of its build task in smith.toml")
	}
	return &protocol.Fix{
		Description: "Build " + p.Name + " into " + target,
		Choices: []protocol.FixChoice{{
			Question: p.Name + " builds into " + path.Join(p.Root, cleanDir(oldOutDir)) + ". smith builds every project into " + target + ". Moving it edits " + strings.Join(files, ", ") + ". Scripts that name the old folder follow it; ignore rules for it stay, harmless. The next screen shows every changed line. The old folder stays on disk and nothing is committed.",
			Options: []protocol.FixChoiceOption{
				{Label: "move the build of " + p.Name + " into " + target, Fix: *move},
				{Label: "leave " + p.Name + " as it is (nothing changes)"},
			},
		}},
	}, ""
}

// MigrationRefusal says why the build cannot be moved safely, or "". It is the
// refusal of doctor and of smith convert --level full: dependents are the
// projects that import p by name.
func MigrationRefusal(workspaceRoot string, p protocol.Project, configName, oldOutDir string, dependents []string) string {
	old := cleanDir(oldOutDir)
	if old == "." || old == "" || strings.HasPrefix(old, "..") {
		return "its build writes into the project folder itself or outside it (" + oldOutDir + "), so there is no build folder to move"
	}
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	raw, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return ""
	}
	var manifest struct {
		Private bool              `json:"private"`
		Scripts map[string]string `json:"scripts"`
	}
	_ = json.Unmarshal(raw, &manifest)
	for _, m := range pathFlag.FindAllStringSubmatch(manifest.Scripts["build"], -1) {
		if !under(m[3], old) {
			return "the build script also writes " + m[3] + ", outside " + old + "; one project needs one build folder"
		}
	}
	for _, o := range p.Tasks["build"].Outputs {
		resolved := path.Clean(path.Join(p.Root, o))
		if !under(resolved, path.Join(p.Root, old)) && !protocol.InBuildDir(resolved) {
			return "smith.toml gives the build task the output " + resolved + ", which is not the build folder; edit it by hand"
		}
	}
	if configName != "" {
		if config, err := os.ReadFile(filepath.Join(dir, configName)); err == nil {
			for _, m := range configPath.FindAllStringSubmatch(string(config), -1) {
				if m[2] != "" && !under(m[2], old) {
					return configName + " also writes " + m[2] + ", outside " + old + "; one project needs one build folder"
				}
			}
		}
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return ""
	}
	if len(dependents) > 0 {
		// Another project reads this one through node_modules, a link: it
		// resolves main and types from the link, not from the project folder,
		// so "../../dist/..." is not found from there and the importers stop
		// compiling.
		pointsIntoBuild := false
		doc.RewriteStrings(buildPathKeys, func(v string) string {
			pointsIntoBuild = pointsIntoBuild || under(v, old)
			return v
		})
		if pointsIntoBuild {
			return "its package.json main, types or exports lead into the build folder and " + strings.Join(dependents, ", ") + " import it by name. Through node_modules, a path outside the package such as ../../dist/... is not found, so they would stop compiling. Keep its build folder, or make the importers read its sources"
		}
	}
	if p.Kind != protocol.ProjectPackage || manifest.Private {
		return ""
	}
	for _, e := range fileEntries(doc) {
		if !under(e, old) && !packagedAnyway(e) {
			return "package.json \"files\" lists " + e + ", which is not in the build folder: smith publishes only dist/" + p.Root + ", so " + e + " would be missing from the tarball"
		}
	}
	reason := ""
	doc.RewriteStrings(buildPathKeys, func(v string) string {
		if reason == "" && v != "./package.json" && !under(v, old) {
			reason = "package.json points at " + v + ", which is not in the build folder: smith publishes only dist/" + p.Root + ", so it would be missing from the tarball"
		}
		return v
	})
	return reason
}

// dropDeclaredOutputs is the step that removes the outputs smith.toml gives
// the build task when they are the old build folder: the plugin's own outputs
// (dist/<project dir>) apply again. It is a smith command, not a rewrite of
// smith.toml, so the steps of several projects never undo each other.
func dropDeclaredOutputs(workspaceRoot string, p protocol.Project, old string) (protocol.FixStep, bool) {
	declared := p.Tasks["build"].Outputs
	if len(declared) == 0 {
		return protocol.FixStep{}, false
	}
	for _, o := range declared {
		if !under(path.Clean(path.Join(p.Root, o)), path.Join(p.Root, old)) {
			return protocol.FixStep{}, false
		}
	}
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "smith.toml"))
	text := string(raw)
	if err != nil || !strings.Contains(text, "projects."+p.Name+".tasks.build]") && !strings.Contains(text, `projects."`+p.Name+`".tasks.build]`) {
		return protocol.FixStep{}, false
	}
	key := p.Name
	if !bareKey.MatchString(key) {
		key = `"` + key + `"`
	}
	return protocol.FixStep{
		Description: "remove outputs = [" + `"` + strings.Join(declared, `", "`) + `"` + "] from [projects." + key + ".tasks.build] in smith.toml, so the build caches " + protocol.ProjectBuildDir(p.Root) + " again",
		Argv:        []string{"smith", "config", "unset", "projects." + key + ".tasks.build.outputs"},
	}, true
}

var bareKey = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// packagedAnyway reports whether npm puts the file in a tarball whatever
// "files" says, and Stage copies it beside the package.
func packagedAnyway(name string) bool {
	upper := strings.ToUpper(path.Base(cleanDir(name)))
	return strings.HasPrefix(upper, "README") || strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "LICENCE") || upper == "PACKAGE.JSON"
}
