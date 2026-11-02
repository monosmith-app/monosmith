package cipipeline

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Preview works out what Write would do and shows it as a plan: every file it
// creates, rewrites or deletes, with a diff of what changes. It writes nothing.
// A file that smith generated but the workspace no longer asks for is listed
// as removed, with the content it had.
//
// Parameters: root is the workspace root; cfg is its configuration; m is the
// pipeline model.
// Returns: the plan (empty when the pipelines are up to date), the slots whose
// content would be lost, and an error when a file cannot be read.
func Preview(root string, cfg workspace.Config, m Model) (operationplan.Plan, []string, error) {
	plan := operationplan.Plan{Title: "generate the CI pipelines"}
	rendered := Rendered(cfg, m)
	var lost []string
	for _, path := range Stale(root, rendered) {
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			return plan, lost, err
		}
		plan.Add("Deleted: smith wrote them and the workspace no longer asks for them", operationplan.Line{
			Action: operationplan.Remove, Path: path, Diff: textdiff.UnifiedDiff(path, normal(string(raw)), ""),
		})
	}
	paths := make([]string, 0, len(rendered))
	for path := range rendered {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		old, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return plan, lost, err
		}
		content, missing := preserveFile(root, path, string(old), rendered[path])
		for _, slot := range missing {
			lost = append(lost, path+": "+slot)
		}
		if string(old) == content {
			continue
		}
		line := operationplan.Line{Action: operationplan.Change, Path: path}
		if err != nil {
			line.Action = operationplan.Create
		} else {
			line.Diff = textdiff.UnifiedDiff(path, normal(string(old)), normal(content))
		}
		plan.Add("Rewritten: smith owns them, edits outside the slots are lost", line)
	}
	return plan, lost, nil
}

// normal turns CRLF into LF so a diff shows content, not line endings.
func normal(s string) string { return strings.ReplaceAll(s, "\r\n", "\n") }
