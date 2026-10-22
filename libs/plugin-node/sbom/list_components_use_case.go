package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the nearest package-lock.json, at the project or any directory
// above it up to the workspace (npm workspaces keep one at the root), and
// falls back to the project's package.json.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "package-lock.json")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		list, err := ComponentsFromLock(lock.Body)
		return protocol.ComponentsResult{Components: list, Source: lock.Rel}, err
	}
	manifest, ok, err := components.ReadIn(root, params.Project.Root, "package.json")
	if err != nil || !ok {
		return protocol.ComponentsResult{}, err
	}
	list, err := ComponentsFromManifest(manifest.Body)
	return protocol.ComponentsResult{Components: list, Source: manifest.Rel}, err
}
