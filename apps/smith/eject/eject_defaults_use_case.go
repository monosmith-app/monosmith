package eject

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Candidate is a tool whose default applies to the workspace right now.
type Candidate struct {
	Tool tooldefaults.Tool
	// Exists is whether the file that would replace the default is already
	// there.
	Exists bool
}

// ResetDefaults deletes what smith wrote under .smith/defaults/ earlier, so
// that after the next discovery only the defaults that apply now are there: a
// file left from before the repository configured the tool would otherwise be
// offered. Plugins write them again whenever they discover projects.
func ResetDefaults(root string) error {
	return os.RemoveAll(filepath.Join(root, filepath.FromSlash(tooldefaults.Dir)))
}

// Available lists the tools whose default file exists under .smith/defaults/,
// in the catalogue's order.
func Available(root string) []Candidate {
	var out []Candidate
	for _, t := range tooldefaults.Tools {
		if _, err := os.Stat(defaultPath(root, t)); err != nil {
			continue
		}
		_, err := os.Stat(targetPath(root, t))
		out = append(out, Candidate{Tool: t, Exists: err == nil})
	}
	return out
}

// Eject copies the tool's default to the file that replaces it, in the
// workspace root, and returns that file's workspace-relative path. It refuses
// to replace a file that is there unless force is set, and says so when the
// tool has no default to copy.
func Eject(root string, tool tooldefaults.Tool, force bool) (string, error) {
	raw, err := os.ReadFile(defaultPath(root, tool))
	if err != nil {
		return "", problem.New(
			"smith has no default for "+tool.Name+" to copy: the repository already configures it, or no project uses it",
			"smith eject lists the defaults that apply",
		)
	}
	target := targetPath(root, tool)
	if have, err := os.ReadFile(target); err == nil {
		if bytes.Equal(have, raw) {
			return tool.Target, nil
		}
		if !force {
			return "", problem.New(tool.Target+" exists already and differs from smith's default", "edit it, or run `smith eject "+tool.Name+" --force` to replace it with the default")
		}
	}
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return "", problem.New("could not write "+tool.Target+": "+err.Error(), "check that the workspace root is writable")
	}
	return tool.Target, nil
}

func defaultPath(root string, t tooldefaults.Tool) string {
	return filepath.Join(root, filepath.FromSlash(tooldefaults.Dir), t.File)
}

func targetPath(root string, t tooldefaults.Tool) string {
	return filepath.Join(root, filepath.FromSlash(t.Target))
}
