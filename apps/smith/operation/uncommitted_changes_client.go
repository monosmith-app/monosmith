package operation

import (
	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
	"github.com/spf13/cobra"
)

// hasChanges reports whether the work tree of root has an uncommitted change
// outside smith's own state. It is false when root is not a git work tree.
func hasChanges(root string) bool {
	git := gitclient.New(root)
	if !git.IsRepository() {
		return false
	}
	status, err := git.Run("status", "--porcelain", "--", ".", exclude)
	return err == nil && status != ""
}

// dryRunning is true for a run that was asked to change nothing.
func dryRunning(c *cobra.Command) bool {
	f := c.Flags().Lookup("dry-run")
	return f != nil && f.Value.String() == "true"
}
