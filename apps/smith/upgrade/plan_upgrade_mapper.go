package upgrade

import "github.com/monosmith-app/monosmith/apps/smith/operationplan"

// PlanUpgrade turns the result of a dry run of smith upgrade into the plan
// smith prints before it writes: the smith.toml migrations, the projects
// brought up to date with their recipe, and the files it rewrites or deletes.
// Owners come from smith.lock where it records a file.
//
// Parameters: root is the workspace root; preview is the Result of the dry run.
func PlanUpgrade(root string, preview Result) operationplan.Plan {
	plan := operationplan.Plan{Title: "upgrade the workspace"}
	for _, m := range preview.Migrations {
		plan.Add("smith.toml", operationplan.Line{Action: operationplan.Change, Path: "smith.toml", Note: m})
	}
	for _, c := range preview.Projects {
		plan.Add("Recipe "+c.Recipe, operationplan.Line{Action: operationplan.Change, Path: c.Project, Note: "brought up to date with the recipe"})
	}
	for _, f := range preview.Files {
		plan.Add("Files smith owns", operationplan.Line{Action: operationplan.Change, Path: f})
	}
	return operationplan.WithOwners(root, plan)
}
