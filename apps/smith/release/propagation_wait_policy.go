package release

import (
	"cmp"
	"errors"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// ErrPropagationPending is returned by a release that finished but did not move
// a dist-tag, because the platform packages were not fetchable from npm in
// time. The report names the command that moves the tag. The command line exits
// with code 2, so CI shows a warning and never a silent success.
var ErrPropagationPending = errors.New("the release is done, but a dist-tag was not moved yet: the report above names the command")

// PropagationWait is the propagation_wait a project's release uses: its own
// table, then [release], as the text of a duration ("" is the default, 15m).
//
// Parameters: cfg is the workspace config; p is the project (nil when unknown).
func PropagationWait(cfg workspace.Config, p *projectgraph.Project) string {
	own := ""
	if p != nil && p.Release != nil {
		own = p.Release.PropagationWait
	}
	return cmp.Or(own, cfg.Release.PropagationWait)
}
