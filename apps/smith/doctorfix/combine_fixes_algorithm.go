package doctorfix

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CombineSafe makes one fix of every planned fix that needs no decision (Ask
// is false): the files of all and each step once, in the order of the plan.
// A fix that writes a file an earlier one writes is left out, because the
// second would be made from content the first replaces; plan again after
// applying. ok is false when there is nothing to combine. A single fix comes
// back as it is.
func CombineSafe(plan []Planned) (combined Planned, ok bool) {
	var taken []Planned
	for _, p := range plan {
		if p.Ask || !disjointFromAll(taken, p) {
			continue
		}
		taken = append(taken, p)
	}
	switch len(taken) {
	case 0:
		return Planned{}, false
	case 1:
		return taken[0], true
	}
	combined.Safe = true
	combined.Description = fmt.Sprintf("%d fixes that need no decision", len(taken))
	for _, p := range taken {
		combined.Files = append(combined.Files, p.Files...)
		for _, s := range p.Steps {
			if !slices.ContainsFunc(combined.Steps, func(o protocol.FixStep) bool { return sameStep(o, s) }) {
				combined.Steps = append(combined.Steps, s)
			}
		}
		for _, c := range p.Checks {
			addCheck(&combined, c)
		}
		addProjects(&combined, p.AppliesTo)
	}
	return combined, true
}

func disjointFromAll(taken []Planned, p Planned) bool {
	for _, t := range taken {
		if !disjoint(t.Files, p.Files) {
			return false
		}
	}
	return true
}
