package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/doctorchecks"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// sharedWorkOutputCheck is the title of the finding on task outputs that hold
// a work folder other projects build in.
const sharedWorkOutputCheck = "no task caches a work folder other projects share"

// sharedWorkOutputFinding turns the lines of SharedWorkOutputs into one
// finding: it passes when there is none, and fails otherwise, with the manual
// step (the outputs are the owner's or the plugin's to change).
func sharedWorkOutputFinding(shared []string) protocol.Finding {
	f := protocol.Finding{Check: sharedWorkOutputCheck, OK: len(shared) == 0, Area: doctorchecks.AreaLint}
	if f.OK {
		return f
	}
	f.Problem = doctorchecks.LeadProblem(shared)
	f.Detail = doctorchecks.Summary(shared, "; ", doctorchecks.ShownProblems) + ". A cache restore empties each output folder first, so it would delete the build state of the other projects"
	f.Remedy = "list only the project's own " + protocol.BuildDir + "/<project dir> in the task's outputs (smith.toml outputs = [...]); for a plugin's own task, update the plugin"
	return f
}
