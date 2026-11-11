package dependencies

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how SwiftPM repairs the packages of the Swift projects. Each
// project is a package with its own Package.swift. It only describes; the host
// deletes and runs.
//
// The kinds:
//   - reinstall: `swift package reset` (it deletes .build, with the checkouts
//     of the dependencies), then `swift package resolve`. Package.resolved
//     keeps the versions.
//   - lockfile: delete Package.resolved, then `swift package resolve` resolves
//     Package.swift again. Offered only when a project has one.
//   - cache: `swift package purge-cache`, then `swift package resolve` in each
//     project. The cache belongs to the user account.
//
// prune is left out: SwiftPM has no command that removes unused dependencies.
// The plugin has no install method, so every plan resolves with its own
// commands.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Swift projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	var reinstall, resolve, lockResolve []protocol.RepairCommand
	var resolved []string
	for _, p := range params.Projects {
		root := path.Clean(filepath.ToSlash(p.Root))
		dir := root
		if dir == "." {
			dir = ""
		}
		reinstall = append(reinstall,
			protocol.RepairCommand{Argv: []string{"swift", "package", "reset"}, Dir: dir},
			protocol.RepairCommand{Argv: []string{"swift", "package", "resolve"}, Dir: dir})
		resolve = append(resolve, protocol.RepairCommand{Argv: []string{"swift", "package", "resolve"}, Dir: dir})
		lock := path.Join(root, "Package.resolved")
		if _, err := os.Lstat(filepath.Join(workspaceRoot, filepath.FromSlash(lock))); err == nil {
			resolved = append(resolved, lock)
			lockResolve = append(lockResolve, protocol.RepairCommand{Argv: []string{"swift", "package", "resolve"}, Dir: dir})
		}
	}
	plans := []protocol.RepairPlan{{
		Kind:     protocol.RepairReinstall,
		Summary:  "swift package reset, then swift package resolve in each package: Package.resolved keeps the versions",
		Commands: reinstall,
	}}
	if len(resolved) > 0 {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete Package.resolved, then swift package resolve resolves Package.swift again",
			Delete:          resolved,
			Commands:        lockResolve,
			ChangesVersions: true,
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear SwiftPM's cache (swift package purge-cache), then swift package resolve",
		Commands:    append([]protocol.RepairCommand{{Argv: []string{"swift", "package", "purge-cache"}}}, resolve...),
		MachineWide: true,
		Notes:       []string{"SwiftPM keeps one cache of repositories and downloads for the user account; every Swift package on this machine fetches again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}
