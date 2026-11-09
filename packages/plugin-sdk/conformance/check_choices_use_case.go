package conformance

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkChoices checks every choice a plugin publishes in its handshake: it is
// well formed, and its name is unique within the plugin. A plugin that
// publishes none passes without a line in the report.
func checkChoices(c *host.Client, r *Report) {
	seen := map[string]bool{}
	for _, choice := range c.Info.Choices {
		err := protocol.ValidateChoice(choice)
		if err == nil && seen[choice.Name] {
			err = fmt.Errorf("the plugin publishes the choice %s twice", choice.Name)
		}
		seen[choice.Name] = true
		r.add("choice "+choice.Name+" is well formed", err == nil, "%v", err)
	}
}
