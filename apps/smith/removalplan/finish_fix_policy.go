package removalplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// The descriptions of the steps that check a workspace, which the screens
// that report their results recognise.
const (
	StepBuild = "build every project"
	StepTest  = "run every project's tests"
	StepLint  = "lint every project"
	// StepFixLint is the lint check when smith owns the lint: ESLint fixes
	// what it can first, then what is left is counted per project.
	StepFixLint = "fix what ESLint can, then lint every project"
	StepDoctor  = "check the workspace with doctor"
)

// Finish is the last step of a full conversion: one npm install refreshes
// package-lock.json for every package taken out and put in, then (unless
// verify is false) the workspace is checked the way a teammate would: it
// builds, tests, lints, and doctor finds nothing wrong. smith is the command
// to run smith with: "smith" to show the plan, the running executable to run
// it. fixLint (the lint choice is smith's) runs ESLint's fixes before the
// lint. The lint is advisory either way: what it finds is reported, and the
// conversion still finishes.
func Finish(smith string, verify, fixLint bool) doctorfix.Planned {
	return finish(smith, verify, verify, fixLint)
}

// FinishBuildAndTest is the last step of a conservative conversion: the same
// install, then the workspace builds, its tests pass and doctor finds nothing
// wrong. Lint is left out: the conservative level keeps the repository's own
// lint setup, which may have complaints that have nothing to do with smith.
func FinishBuildAndTest(smith string) doctorfix.Planned {
	return finish(smith, true, false, false)
}

func finish(smith string, verify, lint, fixLint bool) doctorfix.Planned {
	steps := []protocol.FixStep{
		// One step: smith installs, checks that npm ci accepts the lockfile,
		// installs again when it does not, and rebuilds a git-tracked lockfile
		// as a last resort (lockfilesync).
		{Description: "install the packages and refresh package-lock.json", Argv: []string{smith, "npm-lockfile"}},
	}
	description := "Install the packages and refresh package-lock.json"
	if verify {
		description += ", then check the workspace"
		steps = append(steps,
			protocol.FixStep{Description: StepBuild, Argv: []string{smith, "run", "build"}},
			protocol.FixStep{Description: StepTest, Argv: []string{smith, "run", "test"}},
		)
		if lint {
			step := protocol.FixStep{Description: StepLint, Argv: []string{smith, "run", "lint"}, Advisory: true}
			if fixLint {
				step = protocol.FixStep{Description: StepFixLint, Argv: []string{smith, "eslint-fix"}, Advisory: true}
			}
			steps = append(steps, step)
		}
		steps = append(steps, protocol.FixStep{Description: StepDoctor, Argv: []string{smith, "doctor"}})
	}
	return doctorfix.Planned{
		Fix:    protocol.Fix{Description: description, Safe: true, Steps: steps},
		Checks: []string{"convert --level full: finish"},
	}
}
