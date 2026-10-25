package fullconversion

import (
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
	"github.com/monosmith-app/monosmith/apps/smith/removalplan"
)

// Report is the whole plan of a full conversion as data: what the workspace
// uses, the questions and the choices made, each group of removals as a
// doctor fix, what stays and why, and the closing install and checks. It is
// what `smith convert --level full --json` prints under "full", and what a
// wizard walks.
type Report struct {
	Usage     legacyinventory.Usage     `json:"usage"`
	Choices   convertchoices.Choices    `json:"choices"`
	Questions []convertchoices.Question `json:"questions"`
	Groups    []removalplan.Group       `json:"groups"`
	Kept      []legacyinventory.Kept    `json:"kept,omitempty"`
	Notes     []string                  `json:"notes,omitempty"`
	// Finish installs once and checks the workspace.
	Finish doctorfix.Planned `json:"finish"`
	// Command is the command line that applies this plan without asking.
	Command string `json:"command"`
}

// MakeReport plans the removals for the choices on the converted workspace.
func MakeReport(t *legacyinventory.Tree, c convertchoices.Choices, verify bool) Report {
	usage := legacyinventory.Detect(t)
	plan := removalplan.Build(t, c, nil)
	return Report{
		Usage: usage, Choices: c,
		Questions: convertchoices.Questions(usage, convertchoices.Defaults(usage)),
		Groups:    plan.Groups, Kept: plan.Kept, Notes: plan.Notes,
		Finish:  removalplan.Finish("smith", verify),
		Command: "smith convert " + convertchoices.Flags(c) + " --yes",
	}
}
