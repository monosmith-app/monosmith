package removalplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Finish is the last step of a full conversion: one npm install refreshes
// package-lock.json for every package taken out and put in, then (unless
// verify is false) the workspace is checked the way a teammate would: it
// builds, tests, lints, and doctor finds nothing wrong. smith is the command
// to run smith with: "smith" to show the plan, the running executable to run
// it.
func Finish(smith string, verify bool) doctorfix.Planned {
	steps := []protocol.FixStep{
		{Description: "install the packages and refresh package-lock.json", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}},
		{Description: "install again, so the lockfile settles (npm can leave it one step behind)", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}},
		{Description: "check that npm ci accepts the lockfile", Argv: []string{"npm", "ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund"}},
	}
	description := "Install the packages and refresh package-lock.json"
	if verify {
		description += ", then check the workspace"
		steps = append(steps,
			protocol.FixStep{Description: "build every project", Argv: []string{smith, "run", "build"}},
			protocol.FixStep{Description: "run every project's tests", Argv: []string{smith, "run", "test"}},
			protocol.FixStep{Description: "lint every project", Argv: []string{smith, "run", "lint"}},
			protocol.FixStep{Description: "check the workspace with doctor", Argv: []string{smith, "doctor"}},
		)
	}
	return doctorfix.Planned{
		Fix:    protocol.Fix{Description: description, Safe: true, Steps: steps},
		Checks: []string{"convert --level full: finish"},
	}
}
