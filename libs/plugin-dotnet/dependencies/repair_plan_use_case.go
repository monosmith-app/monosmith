package dependencies

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how NuGet repairs the packages of the .NET projects. It only
// describes; the host deletes and runs.
//
// The kinds:
//   - reinstall: `dotnet restore --force` in each project: every dependency is
//     resolved again, also when the restore looks up to date. The versions stay
//     (a packages.lock.json, or the exact versions of the project file).
//   - lockfile: delete each packages.lock.json, then `dotnet restore
//     --force-evaluate` writes it again. Offered only when a project has one.
//   - cache: `dotnet nuget locals all --clear`, then install (dotnet restore).
//     The NuGet caches belong to the user account.
//
// prune is left out: a restore keeps no package that no project references.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// .NET projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	var force, evaluate []protocol.RepairCommand
	var locks []string
	for _, p := range params.Projects {
		dir := path.Clean(filepath.ToSlash(p.Root))
		if dir == "." {
			dir = ""
		}
		force = append(force, protocol.RepairCommand{Argv: []string{"dotnet", "restore", "--force"}, Dir: dir})
		lock := path.Join(path.Clean(filepath.ToSlash(p.Root)), "packages.lock.json")
		if _, err := os.Lstat(filepath.Join(workspaceRoot, filepath.FromSlash(lock))); err == nil {
			locks = append(locks, lock)
			evaluate = append(evaluate, protocol.RepairCommand{Argv: []string{"dotnet", "restore", "--force-evaluate"}, Dir: dir})
		}
	}
	plans := []protocol.RepairPlan{{
		Kind:     protocol.RepairReinstall,
		Summary:  "dotnet restore --force in each project: NuGet resolves and restores every package again",
		Commands: force,
	}}
	if len(locks) > 0 {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete packages.lock.json, then dotnet restore --force-evaluate writes it again",
			Delete:          locks,
			Commands:        evaluate,
			ChangesVersions: true,
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear the NuGet caches (dotnet nuget locals all --clear), then dotnet restore",
		Commands:    []protocol.RepairCommand{{Argv: []string{"dotnet", "nuget", "locals", "all", "--clear"}}},
		Install:     true,
		MachineWide: true,
		Notes:       []string{"the NuGet caches (global-packages, http-cache, temp, plugins-cache) belong to the user account; every .NET project on this machine restores its packages again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}
