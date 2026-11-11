package dependencies

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// lockfile is the one lockfile of an npm workspace, at the workspace root.
const lockfile = "package-lock.json"

// Repair says how npm repairs the installed dependencies of the projects: one
// plan for each repair kind. It only describes. It reads the disk to name the
// node_modules folders and the lockfile that exist, and it changes nothing: the
// host deletes the paths, runs the commands and then installs.
//
// The four kinds with npm:
//   - prune: `npm prune` removes the packages no package.json declares.
//   - reinstall: delete every node_modules folder, then install. The lockfile
//     stays, so the versions stay.
//   - lockfile: delete every node_modules folder and package-lock.json, then
//     install. npm resolves every range again, so the versions can change. The
//     kind is left out when the workspace has no package-lock.json.
//   - cache: `npm cache clean --force`, then install. The cache belongs to the
//     machine, not to the workspace.
//
// npm installs a workspace as one tree, so the node_modules folder of the
// workspace root is part of every plan that deletes, also when the repair is
// for one project.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Node projects the repair is for.
//
// Returns: the plans, lightest first; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	installed := installedFolders(workspaceRoot, params.Projects)
	plans := []protocol.RepairPlan{
		{
			Kind:     protocol.RepairPrune,
			Summary:  "remove the packages no package.json declares (npm prune)",
			Commands: []protocol.RepairCommand{{Argv: append([]string{"npm", "prune"}, quiet...)}},
		},
		{
			Kind:    protocol.RepairReinstall,
			Summary: "delete every node_modules folder, then npm install: " + lockfile + " keeps the versions",
			Delete:  installed,
			Install: true,
		},
	}
	if exists(filepath.Join(workspaceRoot, lockfile)) {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete every node_modules folder and " + lockfile + ", then npm install: npm resolves every version range again",
			Delete:          append(append([]string(nil), installed...), lockfile),
			Install:         true,
			ChangesVersions: true,
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear the npm download cache (npm cache clean --force), then npm install",
		Commands:    []protocol.RepairCommand{{Argv: []string{"npm", "cache", "clean", "--force"}}},
		Install:     true,
		MachineWide: true,
		Notes:       []string{"npm keeps one cache for the user account (npm config get cache); every workspace on this machine downloads its packages again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}

// installedFolders lists the node_modules folders that exist: the one of the
// workspace root first, then the one of each project, sorted. The paths are
// workspace-relative with forward slashes.
func installedFolders(workspaceRoot string, projects []protocol.Project) []string {
	var out []string
	if exists(filepath.Join(workspaceRoot, "node_modules")) {
		out = append(out, "node_modules")
	}
	seen := map[string]bool{}
	var nested []string
	for _, p := range projects {
		if p.Root == "" || p.Root == "." || seen[p.Root] {
			continue
		}
		seen[p.Root] = true
		if exists(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "node_modules")) {
			nested = append(nested, filepath.ToSlash(p.Root)+"/node_modules")
		}
	}
	sort.Strings(nested)
	return append(out, nested...)
}

// exists reports whether something is at path. It does not follow a link, so
// a node_modules that is a link is still named (and the host deletes the link).
func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}
