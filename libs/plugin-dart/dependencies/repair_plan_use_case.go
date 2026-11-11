package dependencies

import (
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how pub repairs the packages of the Dart and Flutter projects.
// smith resolves the workspace with one `dart pub get` (or `flutter pub get`)
// at the workspace root. It only describes; the host deletes and runs.
//
// The kinds:
//   - reinstall: delete every .dart_tool folder (the package configuration pub
//     writes), then install (pub get). pubspec.lock keeps the versions.
//   - lockfile: delete the .dart_tool folders and every pubspec.lock, then
//     install: pub resolves the version constraints again. Offered only when a
//     pubspec.lock exists.
//   - cache: `dart pub cache clean --force` (or flutter), then install. The pub
//     cache belongs to the user account.
//
// prune is left out: pub has no command that removes unused packages, and
// pub get already leaves out of the package configuration what nothing uses.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// Dart projects.
//
// Returns: the plans; none when params has no project.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	if len(params.Projects) == 0 {
		return protocol.RepairResult{}, nil
	}
	flutter := false
	roots := []string{"."}
	for _, p := range params.Projects {
		if tool(p) == "flutter" {
			flutter = true
		}
		if r := path.Clean(filepath.ToSlash(p.Root)); r != "." {
			roots = append(roots, r)
		}
	}
	sort.Strings(roots[1:])
	var tools, locks []string
	for _, r := range roots {
		for name, into := range map[string]*[]string{".dart_tool": &tools, "pubspec.lock": &locks} {
			rel := path.Join(r, name)
			if _, err := os.Lstat(filepath.Join(workspaceRoot, filepath.FromSlash(rel))); err == nil {
				*into = append(*into, rel)
			}
		}
	}
	program := "dart"
	if flutter {
		program = "flutter"
	}
	plans := []protocol.RepairPlan{{
		Kind:    protocol.RepairReinstall,
		Summary: "delete every .dart_tool folder, then " + program + " pub get: pubspec.lock keeps the versions",
		Delete:  tools,
		Install: true,
	}}
	if len(locks) > 0 {
		plans = append(plans, protocol.RepairPlan{
			Kind:            protocol.RepairLockfile,
			Summary:         "delete the .dart_tool folders and pubspec.lock, then " + program + " pub get resolves every constraint again",
			Delete:          append(append([]string(nil), tools...), locks...),
			Install:         true,
			ChangesVersions: true,
		})
	}
	plans = append(plans, protocol.RepairPlan{
		Kind:        protocol.RepairCache,
		Summary:     "clear the pub cache (" + program + " pub cache clean --force), then " + program + " pub get",
		Commands:    []protocol.RepairCommand{{Argv: []string{program, "pub", "cache", "clean", "--force"}}},
		Install:     true,
		MachineWide: true,
		Notes:       []string{"the pub cache (PUB_CACHE) belongs to the user account; every Dart and Flutter project on this machine downloads its packages again"},
	})
	return protocol.RepairResult{Plans: plans}, nil
}
