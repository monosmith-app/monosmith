package pipeline

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/scaffolding"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckWorkspaceEdition warns when crates repeat their edition instead of
// inheriting it from the root's [workspace.package] table, the way new crates
// do. The fix moves the edition to the root and points the crates at it. It
// is never safe: it edits manifests the user owns, so doctor asks first. It
// reports nothing when no crate sets its own edition.
func CheckWorkspaceEdition(workspaceRoot, rootSrc string, projects []protocol.Project) (protocol.Finding, bool) {
	crates := map[string]string{}
	var names []string
	for _, p := range projects {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "Cargo.toml"))
		if err != nil {
			continue
		}
		crates[p.Root] = string(raw)
		if scaffolding.CrateEdition(string(raw)) != "" {
			names = append(names, p.Name)
		}
	}
	root, changed, edition := scaffolding.ShareEdition(rootSrc, crates)
	if edition == "" {
		return protocol.Finding{}, false
	}
	sort.Strings(names)
	f := protocol.Finding{
		Check: "crates inherit their edition from the workspace", Area: "config", Warning: true,
		Problem: "crates set their own edition, so one edit of the workspace does not reach them",
		Detail:  "these crates set their own edition: " + strings.Join(names, ", ") + ". New crates take it from [workspace.package] (edition.workspace = true)",
		Remedy:  "move the edition to [workspace.package] in the root Cargo.toml and set edition.workspace = true in each crate",
	}
	files := []protocol.FileEdit{{Path: "Cargo.toml", Content: root}}
	dirs := make([]string, 0, len(changed))
	for dir := range changed {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		files = append(files, protocol.FileEdit{Path: filepath.ToSlash(filepath.Join(dir, "Cargo.toml")), Content: changed[dir]})
	}
	f.Fix = &protocol.Fix{Description: "share edition " + edition + " through [workspace.package]", Safe: false, Files: files}
	return f, true
}
