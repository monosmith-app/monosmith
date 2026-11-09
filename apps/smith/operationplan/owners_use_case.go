package operationplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
)

// WithOwners fills Owner on each line whose path smith.lock records. A
// workspace without a record, or with one smith cannot read, gets the plan
// back unchanged: owners are extra information, never a reason to stop.
//
// Parameters: root is the workspace root; plan is the plan to annotate.
func WithOwners(root string, plan Plan) Plan {
	rec, err := derivelock.Read(root)
	if err != nil || len(rec.Files) == 0 {
		return plan
	}
	owners := map[string]string{}
	for _, f := range rec.Files {
		owners[f.Path] = f.Owner
	}
	out := Plan{Title: plan.Title}
	for _, g := range plan.Groups {
		ng := Group{Heading: g.Heading}
		for _, l := range g.Lines {
			if l.Owner == "" {
				l.Owner = owners[l.Path]
			}
			ng.Lines = append(ng.Lines, l)
		}
		out.Groups = append(out.Groups, ng)
	}
	return out
}
