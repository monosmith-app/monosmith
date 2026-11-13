package workspace

import (
	"cmp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Info is the part of the workspace section a plugin may use to name a new
// project: the namespace (the workspace name when none is set), the scope and
// module prefix only when the person set them, and the given module prefix.
//
// Parameters: modulePrefix is the Go module prefix to send: the configured one,
// or the one the git remote gives when the workspace has none.
// Returns: the workspace as the plugin protocol carries it.
func (w WorkspaceSection) Info(modulePrefix string) protocol.WorkspaceInfo {
	return protocol.WorkspaceInfo{
		Name:         w.Name,
		Namespace:    cmp.Or(strings.TrimSpace(w.Namespace), w.Name),
		Scope:        w.Scope,
		ModulePrefix: modulePrefix,
	}
}
