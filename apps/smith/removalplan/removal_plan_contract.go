package removalplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/legacyinventory"
)

// The group IDs, in the order they are asked.
const (
	GroupOutput = "output"
	GroupLint   = "lint"
	GroupTests  = "tests"
	GroupBuild  = "build"
	GroupNx     = "nx"
)

// Group is one set of deletions with the reason that follows from a choice.
// Its fix is the review: files as diffs (a package.json edit is a diff too),
// asked as one question.
type Group struct {
	ID string `json:"id"`
	// Reason is why the group is offered: "you chose smith's ESLint".
	Reason string `json:"reason"`
	// Files and Packages count what the group removes, for the question.
	Files    int `json:"files"`
	Packages int `json:"packages"`
	// Removes name the packages, for a wizard that lists them.
	Removes []string `json:"removes,omitempty"`
	// Fix is what applying the group does, as a doctor fix.
	Fix doctorfix.Planned `json:"fix"`
}

// Plan is the removals of a full conversion, in the order they are asked.
type Plan struct {
	Choices convertchoices.Choices `json:"choices"`
	Groups  []Group                `json:"groups"`
	// Kept are files and packages that look like the old tooling and stay,
	// each with what keeps it.
	Kept []legacyinventory.Kept `json:"kept,omitempty"`
	// Notes are things the choices do not do, said plainly.
	Notes []string `json:"notes,omitempty"`
}
