package eject

import (
	"bytes"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tooldefaults"
)

// Candidate is a tool whose default applies to the workspace right now.
type Candidate struct {
	Tool tooldefaults.Tool
	// Exists is whether the file that would replace the default is already
	// there (always false for a tool that is configured per project).
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
		exists := false
		if !t.PerProject {
			_, err := os.Stat(filepath.Join(root, filepath.FromSlash(t.Target)))
			exists = err == nil
		}
		out = append(out, Candidate{Tool: t, Exists: exists})
	}
	return out
}

// Eject copies the tool's default to the file that replaces it and returns the
// workspace-relative paths written. That is one file in the workspace root, or,
// for a tool that reads its configuration from the project's directory only,
// one in each of projectRoots (the projects that use the default). It refuses
// to replace a file that is there unless force is set, and says so when the
// tool has no default to copy.
func Eject(root string, tool tooldefaults.Tool, force bool, projectRoots []string) ([]string, error) {
	raw, err := os.ReadFile(defaultPath(root, tool))
	if err != nil {
		return nil, problem.New(
			"smith has no default for "+tool.Name+" to copy: the repository already configures it, or no project uses it",
			"smith eject lists the defaults that apply",
		)
	}
	targets := []string{tool.Target}
	if tool.PerProject {
		targets = targets[:0]
		for _, p := range projectRoots {
			targets = append(targets, path.Join(p, tool.Target))
		}
		if len(targets) == 0 {
			return nil, problem.New("no project uses smith's default for "+tool.Name, "every project configures it already")
		}
	}
	// Check every target before writing any, so a refusal changes nothing.
	for _, rel := range targets {
		if have, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))); err == nil && !bytes.Equal(have, raw) && !force {
			return nil, problem.New(rel+" exists already and differs from smith's default", "edit it, or run `smith eject "+tool.Name+" --force` to replace it with the default")
		}
	}
	for _, rel := range targets {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(rel)), raw, 0o644); err != nil {
			return nil, problem.New("could not write "+rel+": "+err.Error(), "check that the directory is writable")
		}
	}
	return targets, nil
}

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

func defaultPath(root string, t tooldefaults.Tool) string {
	return filepath.Join(root, filepath.FromSlash(tooldefaults.Dir), t.File)
}
