package adoptmove

import (
	"fmt"

	"github.com/monosmith-app/monosmith/apps/smith/gitclient"
)

// Rollback puts the repository back as it was before Apply: it drops the staged
// moves (and the .gitignore fix-up) with git reset --hard. The caller checked
// that the work tree was clean before the move, so no work of the person is lost.
// When the move was committed, that commit is dropped too.
//
// Parameters: root is the repository root; committed is true when Commit ran.
// Returns: an error when git cannot reset.
func Rollback(root string, committed bool) error {
	target := "HEAD"
	if committed {
		target = "HEAD~1"
	}
	if _, err := gitclient.New(root).Run("reset", "--hard", "-q", target); err != nil {
		return fmt.Errorf("git reset --hard %s: %w", target, err)
	}
	return nil
}
