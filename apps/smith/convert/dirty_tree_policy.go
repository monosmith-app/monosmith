package convert

import "strings"

// DirtyApartFromInstall says whether git status (--porcelain output) lists a
// change other than what the install smith ran for a dry run leaves behind: the
// untracked .nx folder and package-lock.json. A dry run changes nothing else,
// so it must not make the next convert refuse to run.
//
// Parameters: status is the output of git status --porcelain.
// Returns: true when something else is uncommitted.
func DirtyApartFromInstall(status string) bool {
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if path, untracked := strings.CutPrefix(line, "?? "); untracked {
			switch strings.Trim(path, `"`) {
			case ".nx/", ".nx", "package-lock.json":
				continue
			}
		}
		return true
	}
	return false
}
