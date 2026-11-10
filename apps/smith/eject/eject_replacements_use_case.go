package eject

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Replacement is a file Eject would replace, with what is lost.
type Replacement struct {
	// Path is workspace-relative.
	Path string
	// Diff is the unified diff from the file that is there to smith's default.
	Diff string
}

// Replacements lists the files Eject would replace with the default of a tool,
// each with a diff, so --force never overwrites in the dark. It reads and
// writes nothing else.
//
// Parameters: root is the workspace root; tool is the default; projectRoots
// are the projects that use it (for a per-project tool).
// Returns: the files that exist with other content; none when Eject only
// creates files.
func Replacements(root string, tool tooldefaults.Tool, projectRoots []string) []Replacement {
	raw, err := os.ReadFile(defaultPath(root, tool))
	if err != nil {
		return nil
	}
	targets := []string{tool.Target}
	if tool.PerProject {
		targets = targets[:0]
		for _, p := range projectRoots {
			targets = append(targets, path.Join(p, tool.Target))
		}
	}
	var out []Replacement
	for _, rel := range targets {
		have, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil || bytes.Equal(have, raw) {
			continue
		}
		out = append(out, Replacement{Path: rel, Diff: textdiff.UnifiedDiff(rel, strings.ReplaceAll(string(have), "\r\n", "\n"), strings.ReplaceAll(string(raw), "\r\n", "\n"))})
	}
	return out
}
