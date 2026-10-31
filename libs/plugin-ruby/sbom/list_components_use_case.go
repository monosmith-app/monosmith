package sbom

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// List reads the project's Gemfile.lock, at the project or any directory
// above it up to the workspace, and falls back to its Gemfile and gemspec.
func List(root string, params protocol.ComponentsParams) (protocol.ComponentsResult, error) {
	lock, ok, err := components.FindUp(root, params.Project.Root, "Gemfile.lock")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		return protocol.ComponentsResult{Components: ComponentsFromLock(string(lock.Body)), Source: lock.Rel}, nil
	}
	var texts []string
	source := ""
	gemfile, ok, err := components.ReadIn(root, params.Project.Root, "Gemfile")
	if err != nil {
		return protocol.ComponentsResult{}, err
	}
	if ok {
		texts, source = append(texts, string(gemfile.Body)), gemfile.Rel
	}
	entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(params.Project.Root)))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".gemspec") {
			spec, _, err := components.ReadIn(root, params.Project.Root, e.Name())
			if err != nil {
				return protocol.ComponentsResult{}, err
			}
			texts = append(texts, string(spec.Body))
			if source == "" {
				source = spec.Rel
			}
		}
	}
	return protocol.ComponentsResult{Components: ComponentsFromManifests(texts...), Source: source}, nil
}
