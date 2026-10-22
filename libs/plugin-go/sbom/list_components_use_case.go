package sbom

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's go.mod.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	f, ok, err := components.ReadIn(root, params.Project.Root, "go.mod")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if !ok {
		return protocol.ComponentsResult{}, fmt.Errorf("no go.mod in %s", params.Project.Root)
	}
	res := protocol.ComponentsResult{Components: ComponentsFromGoMod(string(f.Body)), Source: f.Rel}
	if m := ModulePath(string(f.Body)); m != "" {
		res.PURL = "pkg:golang/" + m
	}
	return res, nil
}
