package doctorview

import (
	"fmt"
	"slices"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
)

// Plan is the fixes the failed checks carry, folded as `smith doctor --fix`
// offers them. It is read from the report, or made from its findings when the
// report lacks it.
func (v Model) Plan() []doctorfix.Planned {
	if v.Report == nil {
		return nil
	}
	if len(v.Report.Fixes) > 0 {
		return v.Report.Fixes
	}
	return doctor.Plan(*v.Report)
}

// FixFor is the planned fix that clears a check; ok is false for a check
// without one, which only has its remedy.
func (v Model) FixFor(f doctor.Finding) (doctorfix.Planned, bool) {
	if f.OK {
		return doctorfix.Planned{}, false
	}
	for _, p := range v.Plan() {
		if slices.Contains(p.Checks, f.Check) {
			return p, true
		}
	}
	return doctorfix.Planned{}, false
}

// SelectedFix is the planned fix of the check under the cursor.
func (v Model) SelectedFix() (doctorfix.Planned, bool) {
	f, ok := v.Selected()
	if !ok {
		return doctorfix.Planned{}, false
	}
	return v.FixFor(f)
}

// FixAll is what x and the palette's "Fix with doctor" start on: every fix
// that needs no decision as one, or, when all of them ask, the first of
// those.
func (v Model) FixAll() (doctorfix.Planned, bool) {
	plan := v.Plan()
	if p, ok := doctorfix.CombineSafe(plan); ok {
		return p, true
	}
	if len(plan) > 0 {
		return plan[0], true
	}
	return doctorfix.Planned{}, false
}

// FixLabel is the action that starts a fix: "Fix: <what it does>", or, for a
// fix folded from many projects, "Fix all N projects".
func FixLabel(p doctorfix.Planned) string {
	if n := len(p.AppliesTo); n > 1 {
		return fmt.Sprintf("Fix all %d projects", n)
	}
	return "Fix: " + p.Description
}
