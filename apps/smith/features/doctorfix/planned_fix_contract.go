package doctorfix

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"

// Item is a failed check and the fix it carries; the input of Group.
type Item struct {
	Check string
	Fix   *protocol.Fix
}

// Planned is what doctor --fix offers: one fix, possibly folded from the same
// kind of fix on many projects. It is part of `smith doctor --json` (as
// "fixes") and what a wizard walks through.
type Planned struct {
	protocol.Fix
	// Checks are the failed checks the fix clears.
	Checks []string `json:"checks"`
	// Ask is true when the user must decide: the fix is not Safe, or it has
	// Choices. Only a fix with Ask false is applied by --yes alone.
	Ask bool `json:"ask"`
}

// Asks reports whether a fix needs the user's decision before it runs.
func Asks(f protocol.Fix) bool {
	return !f.Safe || len(f.Choices) > 0
}
