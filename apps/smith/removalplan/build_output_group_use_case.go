package removalplan

import (
	"fmt"
	"path"
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
func buildOutputGroup(v *legacyinventory.Tree) built {
	var out built
	var edits []protocol.FileEdit
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
		fix := manifestfixes.BuildIntoDistFix(v.Root, protocol.Project{Name: name, Root: root}, config, outDir)
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

func projectCount(n int) string {
	if n == 1 {
		return "1 project"
	}
	return fmt.Sprintf("%d projects", n)
}

func pathOnDisk(v *legacyinventory.Tree, rel string) string {
	return v.Root + "/" + rel
}
