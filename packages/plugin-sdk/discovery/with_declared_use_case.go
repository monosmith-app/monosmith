package discovery

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// WithDeclared adds one project to result for each declared project, in
// order, and returns result.
//
// describe turns a declared folder into a project; the plugin owns what that
// means. When describe fails, WithDeclared stops and returns the projects added
// so far with an error that names the declaration, as "[projects.<name>] at
// <root>: <cause>".
func WithDeclared(result protocol.DiscoverResult, declared []protocol.DeclaredProject, describe func(protocol.DeclaredProject) (protocol.Project, error)) (protocol.DiscoverResult, error) {
	for _, d := range declared {
		project, err := describe(d)
		if err != nil {
			return result, fmt.Errorf("[projects.%s] at %s: %w", d.Name, d.Root, err)
		}
		result.Projects = append(result.Projects, project)
	}
	return result, nil
}
