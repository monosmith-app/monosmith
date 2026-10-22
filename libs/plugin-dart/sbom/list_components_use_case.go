package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's pubspec.lock, at the project or any directory
// above it up to the workspace (a pub workspace keeps one at its root), and
// falls back to its pubspec.yaml.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "pubspec.lock")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		list, err := ComponentsFromLock(lock.Body)
		return protocol.ComponentsResult{Components: list, Source: lock.Rel}, err
	}
	manifest, ok, err := components.ReadIn(root, params.Project.Root, "pubspec.yaml")
	if err != nil || !ok {
		return protocol.ComponentsResult{}, err
	}
	list, err := ComponentsFromManifest(manifest.Body)
	return protocol.ComponentsResult{Components: list, Source: manifest.Rel}, err
}
