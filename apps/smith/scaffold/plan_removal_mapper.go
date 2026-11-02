package scaffold

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

// PlanRemoval turns the result of a dry run of smith remove into the plan
// smith prints before it deletes anything: the directory, then the files
// outside it that change, with the owner where smith.lock records one.
//
// Parameters: root is the workspace root; preview is the Removal of a RemoveProject dry run.
func PlanRemoval(root string, preview Removal) operationplan.Plan {
	plan := operationplan.Plan{Title: "remove " + preview.Project}
	note := "the whole directory"
	switch {
	case preview.KeptAt != "":
		note = "moved to " + preview.KeptAt + "/, nothing is lost"
	case len(preview.Unsaved) > 0:
		note = fmt.Sprintf("the whole directory, deleted for good, with %d file(s) git cannot bring back", len(preview.Unsaved))
	}
	plan.Add("Project "+preview.Project, operationplan.Line{Action: operationplan.Remove, Path: preview.Root + "/", Note: note})
	for _, f := range preview.Files {
		plan.Add("Workspace", operationplan.Line{Action: operationplan.Change, Path: f})
	}
	return operationplan.WithOwners(root, plan)
}
