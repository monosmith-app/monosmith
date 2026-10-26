package venv

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreEntry is the .gitignore line that keeps the virtual environment out
// of commits.
const IgnoreEntry = ".venv/"

// GitignoreWith returns a .gitignore's text with IgnoreEntry added, and false
// when the file already lists the venv (.venv, .venv/, /.venv or /.venv/).
func GitignoreWith(current string) (string, bool) {
	for _, line := range strings.Split(current, "\n") {
		switch strings.TrimSpace(line) {
		case ".venv", ".venv/", "/.venv", "/.venv/":
			return current, false
		}
	}
	text := strings.TrimRight(current, "\n")
	if text != "" {
		text += "\n\n"
	}
	return text + "# Python\n" + IgnoreEntry + "\n", true
}

// ignoreVenv makes sure the workspace's .gitignore lists the venv, so the
// environment smith is about to create is never committed. It reports whether
// the file changed.
func ignoreVenv(workspaceRoot string) (bool, error) {
	path := filepath.Join(workspaceRoot, ".gitignore")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	out, changed := GitignoreWith(string(raw))
	if !changed {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(out), 0o644)
}
