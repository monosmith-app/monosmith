package dependencies

import (
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Repair says how Maven repairs the dependencies of the JVM projects. It only
// describes; the host deletes and runs.
//
// One kind has an honest equivalent, for Maven only:
//   - cache: `mvn -B dependency:purge-local-repository`. It deletes the
//     dependencies of the projects from the local repository (~/.m2) and
//     resolves them again. It runs at the workspace root when the root has a
//     pom.xml, else in each Maven project. The local repository belongs to the
//     user account, so another project that uses the same artifacts downloads
//     them again.
//
// The other kinds are left out. Maven and Gradle keep no installed
// dependencies in the workspace (prune and reinstall have nothing to act on),
// and neither keeps a lockfile by default. Gradle has no command that clears
// its cache (--refresh-dependencies checks again, it does not clear), and
// smith never deletes outside the workspace.
//
// Parameters: workspaceRoot is the absolute workspace root; params names the
// JVM projects.
//
// Returns: the plans; none when no project builds with Maven.
func Repair(workspaceRoot string, params protocol.RepairParams) (protocol.RepairResult, error) {
	purge := []string{"mvn", "-B", "dependency:purge-local-repository"}
	var commands []protocol.RepairCommand
	if isFile(filepath.Join(workspaceRoot, "pom.xml")) {
		for _, p := range params.Projects {
			if isFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "pom.xml")) {
				commands = []protocol.RepairCommand{{Argv: purge}}
				break
			}
		}
	} else {
		for _, p := range params.Projects {
			root := path.Clean(filepath.ToSlash(p.Root))
			if isFile(filepath.Join(workspaceRoot, filepath.FromSlash(root), "pom.xml")) {
				commands = append(commands, protocol.RepairCommand{Argv: purge, Dir: root})
			}
		}
	}
	if len(commands) == 0 {
		return protocol.RepairResult{}, nil
	}
	return protocol.RepairResult{Plans: []protocol.RepairPlan{{
		Kind:        protocol.RepairCache,
		Summary:     "delete the dependencies of the Maven projects from the local repository (~/.m2) and resolve them again",
		Commands:    commands,
		MachineWide: true,
		Notes:       []string{"the local repository belongs to the user account; another project that uses the same artifacts downloads them again"},
	}}}, nil
}

// isFile reports whether a regular file is at path.
func isFile(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}
