package conformance

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkTools checks every tool a plugin declares in its handshake: it is well
// formed (it can be detected and installed) and its name is unique within the
// plugin. A plugin that declares none passes without a line in the report.
func checkTools(c *host.Client, r *Report) {
	seen := map[string]bool{}
	for _, tool := range c.Info.Tools {
		err := protocol.ValidateToolNeed(tool)
		if err == nil && seen[tool.Name] {
			err = fmt.Errorf("the plugin declares the tool %s twice", tool.Name)
		}
		seen[tool.Name] = true
		r.add("tool "+tool.Name+" is well formed", err == nil, "%v", err)
	}
}
