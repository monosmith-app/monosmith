package doctorview

import "github.com/monosmith-app/monosmith/apps/smith/doctor"

// runNow are the catalogue actions that run a check of doctor again on its
// own, by the check: its guided form is the one the palette opens.
var runNow = map[string]string{
	doctor.BoundaryRulesCheck: "boundaries",
	doctor.BoundaryAppCheck:   "boundaries",
}

// RunNowAction is the id of the catalogue action that runs a finding's check
// on its own ("boundaries"), or "" when the check has none.
func RunNowAction(f doctor.Finding) string { return runNow[f.Check] }
