package kindchange

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
	"github.com/monosmith-app/monosmith/apps/smith/textdiff"
)

// OperationPlan turns a kind change into the plan smith shows before it
// writes: the folder move, each file it edits with a diff, and the pipelines
// it regenerates.
//
// Parameters: root is the workspace root; p is the worked-out change.
// Returns: the plan to present.
func OperationPlan(root string, p Plan) operationplan.Plan {
	plan := operationplan.Plan{Title: "make " + p.Project + " " + article(p.To) + " (it is " + article(p.From) + ")"}
	if p.Moves() {
		plan.Add("The folder moves with git, so its history follows", operationplan.Line{Action: operationplan.Change, Path: p.Root + " -> " + p.NewRoot, Note: "git mv"})
	}
	for _, e := range p.Edits {
		line := operationplan.Line{Action: operationplan.Change, Path: e.Path}
		switch {
		case e.Delete:
			line.Action = operationplan.Remove
		default:
			before, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(e.Path)))
			if err != nil {
				line.Action, line.Note = operationplan.Create, "new file"
			} else {
				line.Diff = textdiff.UnifiedDiff(e.Path, strings.ReplaceAll(string(before), "\r\n", "\n"), strings.ReplaceAll(e.Content, "\r\n", "\n"))
			}
		}
		plan.Add("Files smith edits", line)
	}
	plan.Add("Regenerated from the new kind", operationplan.Line{Action: operationplan.Change, Path: "the CI pipelines", Note: "smith-owned pipeline files are rewritten; smith ci generate --dry-run shows them"})
	return plan
}

// DirtyPaths lists the paths git status --porcelain reports as changed,
// leaving out smith's own .smith/ folder.
//
// Parameters: status is the output of git status --porcelain.
// Returns: the changed paths; none means the work tree is clean.
func DirtyPaths(status string) []string {
	var paths []string
	for _, line := range strings.Split(status, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path := strings.TrimSpace(line[strings.IndexAny(line, " ")+1:])
		if strings.HasPrefix(path, ".smith/") {
			continue
		}
		paths = append(paths, path)
	}
	return paths
}
