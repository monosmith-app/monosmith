package dependencies

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how Composer repairs the packages of the workspace, which smith
// installs into one vendor/ at the workspace root. It only describes; the host
// deletes and runs.
//
// The kinds:
//   - reinstall: delete vendor/, then install (composer install).
//     composer.lock keeps the versions.
//   - lockfile: delete vendor/ and composer.lock, then install: Composer
//     resolves composer.json again. Offered only when composer.lock exists.
//   - cache: `composer clear-cache`, then install. Composer's cache belongs to
//     the user account.
//
// prune is left out: Composer has no command that removes unused packages;
// composer install already removes from vendor/ what the lockfile does not list.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// PHP projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	var vendor []string
	if _, err := os.Lstat(filepath.Join(workspaceRoot, "vendor")); err == nil {
		vendor = []string{"vendor"}
	}
	plans := []protocol.RepairPlan{{
		Kind:    protocol.RepairReinstall,
		Summary: "delete vendor, then composer install: composer.lock keeps the versions",
		Delete:  vendor,
		Install: true,
	}}
	if _, err := os.Lstat(filepath.Join(workspaceRoot, "composer.lock")); err == nil {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete vendor and composer.lock, then composer resolves composer.json again",
			Delete:          append(append([]string(nil), vendor...), "composer.lock"),
			Install:         true,
			ChangesVersions: true,
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear Composer's cache (composer clear-cache), then composer install",
		Commands:    []protocol.RepairCommand{{Argv: []string{"composer", "clear-cache", "--no-interaction"}}},
		Install:     true,
		MachineWide: true,
		Notes:       []string{"Composer keeps one cache for the user account (composer config cache-dir); every PHP project on this machine downloads its packages again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}
