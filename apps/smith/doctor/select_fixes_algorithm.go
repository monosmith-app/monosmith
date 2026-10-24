package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Only narrows a report to the findings picked by names: a name is a check's
// text as doctor prints it, or the description of the fix a check carries (a
// fix folded from many projects is named by its description). With no names
// it is the report as it is. The counts are not recomputed: it is meant for
// choosing fixes.
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
		if keep[f.Check] || (f.Fix != nil && keep[f.Fix.Description]) {
			findings = append(findings, f)
		}
	}
	r.Findings = findings
	return r
}

// Plan is the fixes the failed findings carry, folded: the same fix on many
// projects is one that says which it applies to.
func Plan(r Report) []doctorfix.Planned {
	var items []doctorfix.Item
	for _, f := range r.Findings {
		if !f.OK && f.Fix != nil {
			items = append(items, doctorfix.Item{Check: f.Check, Fix: f.Fix})
		}
	}
	return doctorfix.Group(items)
}

// Fixes splits the plan into fixes smith may apply on its own, and the others,
// which need the user's decision.
func Fixes(r Report) (safe, manual []protocol.Fix) {
	for _, p := range Plan(r) {
		if p.Ask {
			manual = append(manual, p.Fix)
		} else {
			safe = append(safe, p.Fix)
		}
	}
	return safe, manual
}
