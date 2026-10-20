package doctor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Only narrows a report to the findings whose check is one of names; with no
// names it is the report as it is. The counts are not recomputed: it is meant
// for choosing fixes.
func Only(r Report, names []string) Report {
	if len(names) == 0 {
		return r
	}
	keep := map[string]bool{}
	for _, n := range names {
		keep[n] = true
	}
	var findings []Finding
	for _, f := range r.Findings {
		if keep[f.Check] {
			findings = append(findings, f)
		}
	}
	r.Findings = findings
	return r
}

// Fixes are the distinct fixes the failed findings carry: safe ones smith
// may apply, and the others, which it only shows.
func Fixes(r Report) (safe, manual []protocol.Fix) {
	seen := map[string]bool{}
	for _, f := range r.Findings {
		if f.OK || f.Fix == nil {
			continue
		}
		key := f.Fix.Description
		for _, e := range f.Fix.Files {
			key += "\x00" + e.Path + "\x00" + e.Content
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		if f.Fix.Safe {
			safe = append(safe, *f.Fix)
		} else {
			manual = append(manual, *f.Fix)
		}
	}
	return safe, manual
}

// ShowFix writes a fix as a description and the diff of each file.
func ShowFix(w io.Writer, root string, fix protocol.Fix) error {
	fmt.Fprintf(w, "● %s\n", fix.Description)
	for _, e := range fix.Files {
		path, err := inside(root, e.Path)
		if err != nil {
			return err
		}
		before, _ := os.ReadFile(path)
		after := e.Content
		if e.Delete {
			after = ""
		}
		fmt.Fprint(w, UnifiedDiff(e.Path, string(before), after))
	}
	return nil
}

// ApplyFix writes a fix's files.
func ApplyFix(root string, fix protocol.Fix) error {
	for _, e := range fix.Files {
		path, err := inside(root, e.Path)
		if err != nil {
			return err
		}
		if e.Delete {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(e.Content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// inside resolves a fix's workspace-relative path, refusing one that leaves
// the workspace: a fix comes from a plugin.
func inside(root, rel string) (string, error) {
	path := filepath.Join(root, filepath.FromSlash(rel))
	back, err := filepath.Rel(root, path)
	if err != nil || filepath.IsAbs(rel) || back == ".." || strings.HasPrefix(back, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("a fix may only change files in the workspace, not %s", rel)
	}
	return path, nil
}
