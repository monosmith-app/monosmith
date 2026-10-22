package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's lockfile (uv.lock, poetry.lock), at the project or
// any directory above it up to the workspace, and falls back to the project's
// requirements.txt and pyproject.toml.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "uv.lock", "poetry.lock")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		list, err := ComponentsFromLock(lock.Body)
		return protocol.ComponentsResult{Components: list, Source: lock.Rel}, err
	}
	if req, ok, err := components.ReadIn(root, params.Project.Root, "requirements.txt"); err != nil {
		return protocol.ComponentsResult{}, err
	} else if ok {
		if list := ComponentsFromRequirements(string(req.Body)); len(list) > 0 {
			return protocol.ComponentsResult{Components: list, Source: req.Rel}, nil
		}
	}
	manifest, ok, err := components.ReadIn(root, params.Project.Root, "pyproject.toml")
	if err != nil || !ok {
		return protocol.ComponentsResult{}, err
	}
	list, err := ComponentsFromPyproject(manifest.Body)
	return protocol.ComponentsResult{Components: list, Source: manifest.Rel}, err
}
