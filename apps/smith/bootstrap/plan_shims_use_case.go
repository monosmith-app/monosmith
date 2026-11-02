package bootstrap

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// PlanShims lists the launcher files WriteShims would write: a file that is
// missing is created; one that differs is rewritten, with a diff of what is
// lost. It writes nothing.
//
// Parameters: root is the workspace root.
// Returns: the plan, empty when the launcher is up to date; an error when a
// file cannot be read.
func PlanShims(root string) (operationplan.Plan, error) {
	plan := operationplan.Plan{Title: "write the smith launcher"}
	shims := Shims()
	for _, path := range []string{ShellPath, CmdPath} {
		old, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return plan, err
		}
		if string(old) == shims[path] {
			continue
		}
		line := operationplan.Line{Action: operationplan.Create, Path: path}
		if err == nil {
			line.Action = operationplan.Change
			line.Diff = textdiff.UnifiedDiff(path, strings.ReplaceAll(string(old), "\r\n", "\n"), strings.ReplaceAll(shims[path], "\r\n", "\n"))
		}
		plan.Add("Launcher", line)
	}
	return plan, nil
}
