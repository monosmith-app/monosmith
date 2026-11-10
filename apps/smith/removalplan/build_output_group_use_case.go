package removalplan

import (
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
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
	}
	if projects == 0 {
		return out
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].Path < edits[j].Path })
	reason := fmt.Sprintf("Build %s into the workspace's dist folder: smith passes the folder, so no tsconfig names one and main, types and start follow it", projectCount(projects))
	out.group = newGroup(GroupOutput, reason, 0, nil, edits)
	return out
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
