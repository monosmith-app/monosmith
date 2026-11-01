package adoption

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// existingWorkspaceItem is the blocker for a folder that has a smith.toml. A
// smith.toml that reads fine means there is nothing to adopt. One that cannot
// be read is not "already a workspace" in any useful sense: the blocker names
// the file, line and column the parser reports, and the fix.
func existingWorkspaceItem(raw []byte) Item {
	if _, err := workspace.Parse(raw); err != nil {
		detail := strings.TrimSpace(err.Error())
		return Item{
			Message: "smith.toml cannot be read: " + detail,
			Step:    "fix that line in smith.toml, then run smith adopt again (smith doctor checks it once it reads)",
			Blocks:  BlocksAdoption,
		}
	}
	return Item{Message: "this is already a smith workspace", Step: "nothing to adopt: smith doctor checks it, smith upgrade brings it up to this smith", Blocks: BlocksAdoption}
}
