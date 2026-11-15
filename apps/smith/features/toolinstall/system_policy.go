package toolinstall

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ForSystem keeps the tools that exist on a system: a tool with no Systems
// list is for every system.
//
// Parameters: tools are the declarations; goos is the system as Go names it
// ("windows", "darwin", "linux").
// Returns: the tools for that system, in the same order.
func ForSystem(tools []protocol.ToolNeed, goos string) []protocol.ToolNeed {
	var out []protocol.ToolNeed
	for _, t := range tools {
		if len(t.Systems) == 0 || slices.Contains(t.Systems, goos) {
			out = append(out, t)
		}
	}
	return out
}
