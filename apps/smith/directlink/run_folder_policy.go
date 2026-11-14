package directlink

import (
	"path/filepath"
	"strings"
)

// RunDir is the folder of the hard links smith runs from, on Windows.
func RunDir(home string) string { return filepath.Join(home, "run") }

// BackupDir is the folder of the originals of the entries that were changed.
func BackupDir(home string) string { return filepath.Join(RunDir(home), "backup") }

// StatePath is the file that remembers what was linked.
func StatePath(home string) string { return filepath.Join(home, stateFile) }

// RunningFromLink is true when exe is a link in the run folder: smith already
// starts directly, and the package manager's own file is not the one running.
//
// Parameters: home is the smith home; exe is the running program.
// Returns: true for a program in the run folder.
func RunningFromLink(home, exe string) bool {
	if home == "" || exe == "" {
		return false
	}
	rel, err := filepath.Rel(RunDir(home), exe)
	return err == nil && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel)
}

// InstalledExe is the program the package manager installed: exe itself, or,
// when smith runs from a link, the program that link was made from. The update
// reads how smith was installed from this path.
//
// Parameters: home is the smith home; exe is the running program.
// Returns: the installed program.
func InstalledExe(home, exe string) string {
	if RunningFromLink(home, exe) {
		if s, ok := LoadState(home); ok && s.Installed != "" {
			return s.Installed
		}
	}
	return exe
}
