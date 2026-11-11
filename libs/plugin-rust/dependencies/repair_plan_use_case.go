package dependencies

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how cargo repairs the dependencies of the workspace. It only
// describes; the host deletes and runs.
//
// One kind has an honest equivalent:
//   - lockfile: delete Cargo.lock, then install (`cargo fetch` resolves the
//     version requirements again and writes Cargo.lock). Offered only when
//     Cargo.lock exists.
//
// The other kinds are left out. Cargo keeps no installed dependencies in the
// workspace: the sources live in the machine's cargo registry, and target/ is
// build output, not dependencies, so prune and reinstall have nothing to act
// on. Cargo has no stable command that clears its registry cache (cargo clean
// gc is unstable), and smith never deletes outside the workspace.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Rust projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	if _, err := os.Lstat(filepath.Join(workspaceRoot, "Cargo.lock")); err != nil {
		return protocol.RepairResult{}, nil
	}
	return protocol.RepairResult{Plans: []protocol.RepairPlan{{
		Kind:            protocol.RepairLockfile,
		Summary:         "delete Cargo.lock, then cargo fetch resolves every version requirement again",
		Delete:          []string{"Cargo.lock"},
		Install:         true,
		ChangesVersions: true,
	}}}, nil
}
