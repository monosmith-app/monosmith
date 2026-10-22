package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's Cargo.lock, at the crate or any directory above it
// up to the workspace (a cargo workspace keeps one at its root), and falls
// back to the crate's Cargo.toml.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "Cargo.lock")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		list, err := ComponentsFromLock(lock.Body)
		return protocol.ComponentsResult{Components: list, Source: lock.Rel}, err
	}
	manifest, ok, err := components.ReadIn(root, params.Project.Root, "Cargo.toml")
	if err != nil || !ok {
		return protocol.ComponentsResult{}, err
	}
	list, err := ComponentsFromManifest(manifest.Body)
	return protocol.ComponentsResult{Components: list, Source: manifest.Rel}, err
}
