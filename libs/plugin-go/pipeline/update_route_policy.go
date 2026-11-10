package pipeline

import (
	"os"
	"path/filepath"
	"strings"
)

// updateRoute decides how a found copy of a tool is updated in place. It
// returns the command and the environment that update the copy at path to the
// pinned version, and ok = false when the way the copy was installed is not
// known. It never guesses: an unknown origin is not updated.
type updateRoute func(root, path string) (argv, env []string, ok bool)

// updateRoutes holds the route of each tool that has one, by program name. A
// tool joins by adding its route here.
var updateRoutes = map[string]updateRoute{lintProgram: goInstallRoute}

// unknownOrigin is the reason shown for a copy that has no known route.
const unknownOrigin = "smith does not know how this copy was installed"

// goInstallRoute knows a copy that sits in a Go bin folder (GOBIN, a GOPATH
// entry's bin, or ~/go/bin) and that was built by Go (its build info is
// readable). It updates the copy in place: the same install command that CI
// uses, with GOBIN set to the folder of the copy, so the file that is replaced
// is the one that was found.
func goInstallRoute(root, path string) (argv, env []string, ok bool) {
	dir := filepath.Dir(path)
	inGoBin := false
	for _, bin := range goBinFolders(root) {
		if sameFile(bin, dir) {
			inGoBin = true
			break
		}
	}
	if !inGoBin {
		return nil, nil, false
	}
	// A program that go cannot read build info from was not built by Go.
	out, err := probe(root, []string{"go", "version", "-m", path})
	if err != nil || !strings.Contains(out, "\tpath\t") {
		return nil, nil, false
	}
	return LintInstallCommand(), []string{"GOBIN=" + dir}, true
}

// goBinFolders lists the folders where a Go program is normally installed:
// GOBIN, the bin folder of each GOPATH entry, and ~/go/bin. A folder that go
// does not report is left out.
func goBinFolders(root string) []string {
	var folders []string
	out, err := probe(root, []string{"go", "env", "GOBIN", "GOPATH"})
	if lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(out), "\r", ""), "\n"); err == nil && len(lines) >= 2 {
		if bin := strings.TrimSpace(lines[0]); bin != "" {
			folders = append(folders, bin)
		}
		for _, entry := range filepath.SplitList(strings.TrimSpace(lines[1])) {
			if entry != "" {
				folders = append(folders, filepath.Join(entry, "bin"))
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		folders = append(folders, filepath.Join(home, "go", "bin"))
	}
	return folders
}
