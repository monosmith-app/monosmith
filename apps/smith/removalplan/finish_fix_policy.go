package removalplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The descriptions of the steps that check a workspace, which the screens
// that report their results recognise.
const (
	StepBuild  = "build every project"
	StepTest   = "run every project's tests"
	StepLint   = "lint every project"
	StepDoctor = "check the workspace with doctor"
)

// Finish is the last step of a full conversion: one npm install refreshes
// package-lock.json for every package taken out and put in, then (unless
// verify is false) the workspace is checked the way a teammate would: it
// builds, tests, lints, and doctor finds nothing wrong. smith is the command
// to run smith with: "smith" to show the plan, the running executable to run
// it.
func Finish(smith string, verify bool) doctorfix.Planned {
	return finish(smith, verify, verify)
}

// FinishBuildAndTest is the last step of a conservative conversion: the same
// install, then the workspace builds, its tests pass and doctor finds nothing
// wrong. Lint is left out: the conservative level keeps the repository's own
// lint setup, which may have complaints that have nothing to do with smith.
func FinishBuildAndTest(smith string) doctorfix.Planned {
	return finish(smith, true, false)
}

func finish(smith string, verify, lint bool) doctorfix.Planned {
	steps := []protocol.FixStep{
		{Description: "install the packages and refresh package-lock.json", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}},
		{Description: "install again so the lockfile settles (npm can leave it one step behind)", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}},
		{Description: "check that a clean install from the lockfile works", Argv: []string{"npm", "ci", "--dry-run", "--ignore-scripts", "--no-audit", "--no-fund"}},
	}
	description := "Install the packages and refresh package-lock.json"
	if verify {
		description += ", then check the workspace"
		steps = append(steps,
			protocol.FixStep{Description: StepBuild, Argv: []string{smith, "run", "build"}},
			protocol.FixStep{Description: StepTest, Argv: []string{smith, "run", "test"}},
		)
		if lint {
			steps = append(steps, protocol.FixStep{Description: StepLint, Argv: []string{smith, "run", "lint"}})
		}
		steps = append(steps, protocol.FixStep{Description: StepDoctor, Argv: []string{smith, "doctor"}})
	}
	return doctorfix.Planned{
		Fix:    protocol.Fix{Description: description, Safe: true, Steps: steps},
		Checks: []string{"convert --level full: finish"},
	}
}
