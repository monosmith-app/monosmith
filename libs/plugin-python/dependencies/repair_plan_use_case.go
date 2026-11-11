package dependencies

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-python/venv"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how pip repairs the Python environment of the workspace: the one
// .venv at the workspace root that every Python project installs into. It
// only describes; the host deletes and runs.
//
// Two kinds have an honest equivalent:
//   - reinstall: delete .venv, then install (smith creates .venv again and
//     installs each project in editable mode). smith keeps no lockfile for pip,
//     so pip resolves the version ranges again: the versions can change.
//   - cache: `python -m pip cache purge` with the interpreter of .venv, then
//     install. pip's cache belongs to the user account. Offered only while
//     .venv exists.
//
// prune and lockfile are left out: pip has no command that removes the
// packages nothing requires, and smith keeps no lockfile for pip.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Python projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	reinstall := protocol.RepairPlan{
		Kind:            protocol.RepairReinstall,
		Summary:         "delete .venv, then create it again and install every project",
		Install:         true,
		ChangesVersions: true,
		Notes:           []string{"smith keeps no lockfile for pip, so pip resolves the version ranges of pyproject.toml again"},
	}
	if _, err := os.Lstat(filepath.Join(workspaceRoot, ".venv")); err == nil {
		reinstall.Delete = []string{".venv"}
	}
	plans := []protocol.RepairPlan{reinstall}
	python := venv.Path(workspaceRoot)
	if _, err := os.Stat(python); err == nil {
		plans = append(plans, protocol.RepairPlan{
			Kind:        protocol.RepairCache,
			Summary:     "clear pip's download cache (pip cache purge), then install every project",
			Commands:    []protocol.RepairCommand{{Argv: []string{python, "-m", "pip", "cache", "purge"}}},
			Install:     true,
			MachineWide: true,
			Notes:       []string{"pip keeps one cache for the user account (pip cache dir); every Python environment on this machine downloads its packages again"},
		})
	}
	return protocol.RepairResult{Plans: plans}, nil
}
