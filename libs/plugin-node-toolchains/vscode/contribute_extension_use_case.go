package vscode

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// contributeExtensions marks the extension apps among the projects as ones this
// plugin ships: a release publishes them with it. The node plugin already gives
// them a package task (the staged build, packed into a .vsix in dist/drop); the
// contribution only says that task needs the build, which is also what makes
// the project this plugin's to deploy. A project that is not an extension app
// gets nothing.
func contributeExtensions(workspaceRoot string, params protocol.ContributeParams) protocol.ContributeResult {
	result := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		if isExtensionApp(workspaceRoot, p) {
			result.Projects[p.Name] = protocol.Contribution{DependsOn: map[string][]string{"package": {"build"}}}
		}
	}
	return result
}
