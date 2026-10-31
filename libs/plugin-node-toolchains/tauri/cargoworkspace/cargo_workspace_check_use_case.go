package cargoworkspace

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tauri/tauriapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var (
	membersList = regexp.MustCompile(`(?s)\bmembers\s*=\s*\[(.*?)\]`)
	excludeList = regexp.MustCompile(`(?s)\bexclude\s*=\s*\[(.*?)\]`)
	quoted      = regexp.MustCompile(`"([^"]*)"`)
)

// manifestList returns the strings of a `members` or `exclude` array.
func manifestList(list *regexp.Regexp, manifest string) []string {
	var out []string
	if m := list.FindStringSubmatch(manifest); m != nil {
		for _, q := range quoted.FindAllStringSubmatch(m[1], -1) {
			out = append(out, strings.Trim(q[1], "/"))
		}
	}
	return out
}

// workspaceMembers returns the src-tauri folders (workspace-relative) of the
// Tauri projects at roots that the root Cargo.toml's [workspace] members glob
// matches and its exclude does not.
func workspaceMembers(workspaceRoot string, roots []string) []string {
	src, err := os.ReadFile(filepath.Join(workspaceRoot, "Cargo.toml"))
	if err != nil || !strings.Contains(string(src), "[workspace]") {
		return nil
	}
	members, excluded := manifestList(membersList, string(src)), manifestList(excludeList, string(src))
	var hits []string
	for _, root := range roots {
		dir := path.Join(root, tauriapp.SrcDir)
		if !matchesAny(excluded, dir) && matchesAny(members, dir) {
			hits = append(hits, dir)
		}
	}
	return hits
}

func matchesAny(globs []string, dir string) bool {
	for _, g := range globs {
		if ok, _ := path.Match(g, dir); ok {
			return true
		}
	}
	return false
}

// WorkspaceFinding warns when a root Cargo workspace would take src-tauri in:
// the app has its own Cargo.lock and builds into dist/, and as a member it
// would share the workspace's lockfile and target directory instead.
func WorkspaceFinding(workspaceRoot string, roots []string) protocol.Finding {
	hits := workspaceMembers(workspaceRoot, roots)
	return protocol.Finding{
		Check:   "src-tauri is outside the root Cargo workspace",
		OK:      len(hits) == 0,
		Warning: true,
		Detail:  "the root Cargo.toml [workspace] members match " + strings.Join(hits, ", "),
		Remedy:  "narrow members so they do not match src-tauri, or list it under exclude",
	}
}
