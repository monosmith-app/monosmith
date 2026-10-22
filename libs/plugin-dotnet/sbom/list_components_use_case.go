package sbom

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's packages.lock.json and falls back to the
// PackageReference items of the project file in its directory.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.ReadIn(root, params.Project.Root, "packages.lock.json")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		list, err := ComponentsFromLock(lock.Body)
		return protocol.ComponentsResult{Components: list, Source: lock.Rel}, err
	}
	dir := filepath.Join(root, filepath.FromSlash(params.Project.Root))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	for _, e := range entries {
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".csproj", ".fsproj", ".vbproj":
			f, _, err := components.ReadIn(root, params.Project.Root, e.Name())
			if err != nil {
				return protocol.ComponentsResult{}, err
			}
			list, err := ComponentsFromProject(f.Body)
			return protocol.ComponentsResult{Components: list, Source: path.Clean(f.Rel)}, err
		}
	}
	return protocol.ComponentsResult{}, nil
}
