package dependencies

import (
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how Bundler repairs the gems of the workspace, which smith
// installs with one `bundle install` at the workspace root. It only describes;
// the host deletes and runs.
//
// The kinds:
//   - reinstall: `bundle install --redownload`: every gem is downloaded and
//     installed again. Gemfile.lock keeps the versions.
//   - lockfile: delete Gemfile.lock, then install (bundle install resolves the
//     Gemfile again). Offered only when Gemfile.lock exists.
//
// prune is left out: `bundle clean` removes gems only from a bundle path the
// project configured, and with --force it removes gems of the whole Ruby
// installation, which other projects use. cache is left out: Bundler and
// RubyGems have no command that clears their download cache.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Ruby projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	plans := []protocol.RepairPlan{{
		Kind:     protocol.RepairReinstall,
		Summary:  "bundle install --redownload: download and install every gem again",
		Commands: []protocol.RepairCommand{{Argv: []string{"bundle", "install", "--redownload"}}},
	}}
	if _, err := os.Lstat(filepath.Join(workspaceRoot, "Gemfile.lock")); err == nil {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete Gemfile.lock, then bundle install resolves the Gemfile again",
			Delete:          []string{"Gemfile.lock"},
			Install:         true,
			ChangesVersions: true,
		})
	}
	return protocol.RepairResult{Plans: plans}, nil
}
