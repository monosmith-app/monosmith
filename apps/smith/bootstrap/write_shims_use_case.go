package bootstrap

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func crlf(s string) string { return strings.ReplaceAll(s, "\n", "\r\n") }

// Shims are the launcher files by workspace-relative path.
func Shims() map[string]string {
	return map[string]string{ShellPath: ShellShim(), CmdPath: CmdShim()}
}

// WriteShims writes the launcher files that are missing or different. The sh
// launcher is made executable (where the OS has modes). A launcher the person
// edited is rewritten: the files are smith's, like the CI install script.
//
// Parameters: root is the workspace root; check makes it write nothing.
// Returns: the workspace-relative paths that changed (or would), sorted as
// ShellPath then CmdPath; an error when a file cannot be read or written.
func WriteShims(root string, check bool) ([]string, error) {
	var changed []string
	shims := Shims()
	for _, path := range []string{ShellPath, CmdPath} {
		full := filepath.Join(root, filepath.FromSlash(path))
		old, err := os.ReadFile(full)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return changed, err
		}
		if string(old) == shims[path] {
			continue
		}
		changed = append(changed, path)
		if check {
			continue
		}
		mode := os.FileMode(0o644)
		if path == ShellPath {
			mode = 0o755
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return changed, err
		}
		if err := os.WriteFile(full, []byte(shims[path]), mode); err != nil {
			return changed, err
		}
	}
	return changed, nil
}
