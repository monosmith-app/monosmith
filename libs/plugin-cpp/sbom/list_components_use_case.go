package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's vcpkg.json and conanfile.txt, in its directory or
// any above it up to the workspace. Neither is not an error: the project
// declares no third-party library and its SBOM says so with an empty list.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	var res protocol.ComponentsResult
	if f, ok, err := components.FindUp(root, params.Project.Root, "vcpkg.json"); err != nil {
		return res, err
	} else if ok {
		list, err := ComponentsFromVcpkg(f.Body)
		if err != nil {
			return res, err
		}
		res.Components, res.Source = list, f.Rel
	}
	if f, ok, err := components.FindUp(root, params.Project.Root, "conanfile.txt"); err != nil {
		return res, err
	} else if ok {
		res.Components = components.Dedupe(append(res.Components, ComponentsFromConan(string(f.Body))...))
		if res.Source == "" {
			res.Source = f.Rel
		}
	}
	return res, nil
}
