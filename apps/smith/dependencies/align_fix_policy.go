package dependencies

import (
	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// alignFix is the fix of a one-version check that failed: `smith deps up
// --align`, shown with its plan before it writes.
//
// Parameters: failed is true when the check failed.
// Returns: the fix, or nil when the check passed.
func alignFix(failed bool) *protocol.Fix {
	if !failed {
		return nil
	}
	return smithfix.Fix("Bring every dependency to one version",
		smithfix.Step("Bring each dependency declared at different versions to one version", "deps", "up", "--align"))
}
