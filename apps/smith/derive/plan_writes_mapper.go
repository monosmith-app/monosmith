package derive

import (
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

// PlanWrites is what smith derive will write for a plan: each file smith
// creates or rewrites, grouped by the renderer that produces it, with its owner
// from the record, and the record itself. A file that needs a choice (drift,
// diverged, broken) is left out: derive holds it back.
//
// Parameters: root is the workspace root; plan is the derive plan.
func PlanWrites(root string, plan Plan) operationplan.Plan {
	out := operationplan.Plan{Title: "write the derived files"}
	for _, i := range plan.Items {
		var action operationplan.Action
		switch i.Status {
		case StatusMissing:
			action = operationplan.Create
		case StatusStale:
			action = operationplan.Change
		default:
			continue
		}
		out.Add("Rendered by "+i.File.Renderer+" (plugin "+i.Plugin+")", operationplan.Line{Action: action, Path: i.File.Path, Note: i.Detail, Owner: i.Owner})
	}
	if !out.Empty() {
		out.Add("Record", operationplan.Line{Action: operationplan.Change, Path: derivelock.FileName})
	}
	return out
}
