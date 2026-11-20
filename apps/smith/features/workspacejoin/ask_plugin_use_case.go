package workspacejoin

import (
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Ask asks one language plugin whether its projects are joined to the
// language's workspace and returns the findings. A plugin without the join
// capability, or with no projects, gives none.
//
// Parameters: c is the plugin; projects are the projects it discovered.
// Returns: the findings, or the error of the call.
func Ask(c *host.Client, projects []protocol.Project) ([]protocol.Finding, error) {
	if len(projects) == 0 || !c.Supports(protocol.MethodJoin) {
		return nil, nil
	}
	var res protocol.JoinResult
	if err := c.Call(protocol.MethodJoin, protocol.JoinParams{Projects: projects}, &res, time.Minute); err != nil {
		return nil, err
	}
	return Findings(projects, res), nil
}
