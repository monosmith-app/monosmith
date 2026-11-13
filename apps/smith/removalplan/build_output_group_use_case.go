package removalplan

import (
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/buildoutputs"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	smithconfig "github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// buildOutputGroup moves the output of every project that compiles with tsc
// into the workspace's one dist folder: the folder leaves the tsconfig, the
// build script passes it, and main, types and start follow. It deletes
// nothing; it is the part of embracing smith's build that edits.
//
// A project that doctor refuses to move stays where it is, with the same
// reason (another project imports it by name and its entry points lead into
// the old folder, a published package needs a file outside the build folder):
// the reason is a note of the review.
func buildOutputGroup(v *legacyinventory.Tree) built {
	var out built
	var edits []protocol.FileEdit
	var moved []movedBuild
	importers := importersByName(v)
	projects := 0
	for _, manifest := range v.ManifestPaths() {
		if manifest == "package.json" {
			continue
		}
		root := path.Dir(manifest)
		doc, err := jsondoc.ParseDocument([]byte(v.Read(manifest)))
		if err != nil {
			continue
		}
		scripts := doc.Object("scripts", false)
		if scripts == nil {
			continue
		}
		if build, _ := scripts.String("build"); !strings.Contains(build, "tsc") {
			continue
		}
		name, _ := doc.String("name")
		config, outDir, ok := manifestfixes.LocateOutDir(pathOnDisk(v, root))
		if !ok || protocol.InBuildDir(path.Clean(path.Join(root, outDir))) {
			continue
		}
		project := protocol.Project{Name: name, Root: root, Exports: []string{name}}
		var flags struct{ Private bool }
		_ = json.Unmarshal([]byte(v.Read(manifest)), &flags)
		if !flags.Private {
			project.Kind = protocol.ProjectPackage
		}
		if reason := manifestfixes.MigrationRefusal(v.Root, project, config, outDir, importers[name]); reason != "" {
			out.notes = append(out.notes, name+" ("+root+") stays unmoved: "+reason)
			continue
		}
		fix := manifestfixes.BuildIntoDistFix(v.Root, project, config, outDir)
		if fix == nil {
			continue
		}
		projects++
		edits = append(edits, fix.Files...)
		moved = append(moved, movedBuild{root: root, oldDir: outDir, script: editedBuildScript(fix.Files, manifest)})
	}
	if projects == 0 {
		return out
	}
	reason := fmt.Sprintf("Build %s into the workspace's dist folder: smith passes the folder, so no tsconfig names one and main, types and start follow it", projectCount(projects))
	if config, ok := followOutputs(v.Read(smithconfig.FileName), moved); ok {
		edits = append(edits, protocol.FileEdit{Path: smithconfig.FileName, Content: config})
		reason += ", and so do the build outputs in " + smithconfig.FileName
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].Path < edits[j].Path })
	out.group = newGroup(GroupOutput, reason, 0, nil, edits)
	return out
}

// movedBuild is a project whose build the output group moves: its root, the
// folder its build wrote, and its build script once moved.
type movedBuild struct {
	root, oldDir, script string
}

// editedBuildScript reads the build script of the manifest the fix rewrites.
func editedBuildScript(files []protocol.FileEdit, manifest string) string {
	for _, f := range files {
		if f.Path != manifest {
			continue
		}
		var m struct{ Scripts map[string]string }
		if json.Unmarshal([]byte(f.Content), &m) == nil {
			return m.Scripts[buildoutputs.Task]
		}
	}
	return ""
}

// followOutputs moves the declared build outputs of each moved project to
// the folder its build writes now, with the detector smith doctor uses: the
// declared output that held the old folder is replaced, the others stay. A
// build task that declares no outputs, or runs another script, is left as it
// is (the plugin's default is the workspace's dist folder already).
//
// Parameters: config is the smith.toml text ("" when there is none); moved
// are the projects the group moves.
// Returns: the new smith.toml text, and false when nothing changes.
func followOutputs(config string, moved []movedBuild) (string, bool) {
	if config == "" {
		return "", false
	}
	cfg, err := smithconfig.Parse([]byte(config))
	if err != nil {
		return "", false
	}
	var gaps []buildoutputs.Gap
	for _, m := range moved {
		name := path.Base(m.root)
		task, ok := cfg.Projects[name].Tasks[buildoutputs.Task]
		if !ok || len(task.Outputs) == 0 {
			continue
		}
		if script, _, runs := buildoutputs.RunScript(task.Command); len(task.Command) > 0 && (!runs || script != buildoutputs.Task) {
			continue
		}
		written := buildoutputs.WrittenDirs(m.script)
		if len(written) == 0 {
			continue
		}
		old := path.Clean(m.oldDir)
		var outputs []string
		for _, d := range task.Outputs {
			if d = path.Clean(d); !buildoutputs.Covered(old, []string{d}) && !buildoutputs.Covered(d, []string{old}) {
				outputs = append(outputs, d)
			}
		}
		outputs = buildoutputs.Collapse(append(written, outputs...))
		gaps = append(gaps, buildoutputs.Gap{Project: name, Outputs: outputs})
	}
	if len(gaps) == 0 {
		return "", false
	}
	return buildoutputs.DeclareOutputs(config, gaps), true
}

// importersByName maps a package name to the other packages of the tree that
// list it as a dependency.
func importersByName(v *legacyinventory.Tree) map[string][]string {
	importers := map[string][]string{}
	for _, manifest := range v.ManifestPaths() {
		var m struct {
			Name                                                                  string
			Dependencies, DevDependencies, PeerDependencies, OptionalDependencies map[string]string
		}
		if json.Unmarshal([]byte(v.Read(manifest)), &m) != nil || m.Name == "" {
			continue
		}
		for _, deps := range []map[string]string{m.Dependencies, m.DevDependencies, m.PeerDependencies, m.OptionalDependencies} {
			for dep := range deps {
				if !slices.Contains(importers[dep], m.Name) {
					importers[dep] = append(importers[dep], m.Name)
				}
			}
		}
	}
	return importers
}

func projectCount(n int) string {
	if n == 1 {
		return "1 project"
	}
	return fmt.Sprintf("%d projects", n)
}

func pathOnDisk(v *legacyinventory.Tree, rel string) string {
	return v.Root + "/" + rel
}
