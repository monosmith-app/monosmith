package dependencies

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how the go command repairs the modules of the projects. Go keeps
// no installed dependencies in the workspace: the module cache belongs to the
// machine. It only describes; the host deletes and runs.
//
// The kinds:
//   - prune: `go mod tidy` in each module. It removes the requirements no
//     package imports, and adds one the code imports but go.mod lacks, so a
//     version can be added.
//   - lockfile: delete each go.sum (and go.work.sum), then `go mod tidy` writes
//     go.sum again. go.mod holds the versions; tidy can still add a missing
//     requirement. Offered only when a go.sum exists.
//   - cache: `go clean -modcache`, then install (`go mod download` in each
//     module). The module cache belongs to the machine.
//
// reinstall is left out: there is nothing installed in the workspace to delete.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Go projects (each is a module).
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	var tidy []protocol.RepairCommand
	var sums []string
	for _, p := range params.Projects {
		root := path.Clean(filepath.ToSlash(p.Root))
		if _, err := os.Stat(filepath.Join(workspaceRoot, filepath.FromSlash(root), "go.mod")); err != nil {
			continue
		}
		dir := root
		if dir == "." {
			dir = ""
		}
		tidy = append(tidy, protocol.RepairCommand{Argv: []string{"go", "mod", "tidy"}, Dir: dir})
		sum := path.Join(root, "go.sum")
		if _, err := os.Lstat(filepath.Join(workspaceRoot, filepath.FromSlash(sum))); err == nil {
			sums = append(sums, sum)
		}
	}
	plans := []protocol.RepairPlan{}
	if len(tidy) > 0 {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairPrune,
			Summary:         "go mod tidy in each module: drop the requirements nothing imports",
			Commands:        tidy,
			ChangesVersions: true,
			Notes:           []string{"go mod tidy also adds a requirement the code imports but go.mod lacks, and edits go.mod"},
		})
	}
	if len(sums) > 0 {
		if _, err := os.Lstat(filepath.Join(workspaceRoot, "go.work.sum")); err == nil {
			sums = append(sums, "go.work.sum")
		}
		plans = append(plans, protocol.RepairPlan{
			Kind:     protocol.RepairLockfile,
			Summary:  "delete go.sum, then go mod tidy writes it again from go.mod",
			Delete:   sums,
			Commands: tidy,
			Notes:    []string{"go.mod keeps the versions; go mod tidy can still add a requirement go.mod lacks"},
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear the Go module cache (go clean -modcache), then go mod download",
		Commands:    []protocol.RepairCommand{{Argv: []string{"go", "clean", "-modcache"}}},
		Install:     true,
		MachineWide: true,
		Notes:       []string{"the module cache (go env GOMODCACHE) belongs to the machine; every Go module on it downloads again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}
