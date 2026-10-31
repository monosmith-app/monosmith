package scaffold

import "github.com/monosmith-app/monosmith/apps/smith/operationplan"

// PlanRemoval turns the result of a dry run of smith remove into the plan
// smith prints before it deletes anything: the directory, then the files
// outside it that change, with the owner where smith.lock records one.
//
// Parameters: root is the workspace root; preview is the Removal of a RemoveProject dry run.
func PlanRemoval(root string, preview Removal) operationplan.Plan {
	plan := operationplan.Plan{Title: "remove " + preview.Project}
	plan.Add("Project "+preview.Project, operationplan.Line{Action: operationplan.Remove, Path: preview.Root + "/", Note: "the whole directory"})
	for _, f := range preview.Files {
		plan.Add("Workspace", operationplan.Line{Action: operationplan.Change, Path: f})
	}
	return operationplan.WithOwners(root, plan)
}
