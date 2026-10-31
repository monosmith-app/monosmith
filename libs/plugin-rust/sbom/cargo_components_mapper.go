package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/pelletier/go-toml/v2"
)

// ComponentsFromLock lists the registry and git packages a Cargo.lock
// pins. Packages without a source are the workspace's own crates.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Package []struct {
			Name    string `toml:"name"`
			Version string `toml:"version"`
			Source  string `toml:"source"`
		} `toml:"package"`
	}
	if err := toml.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for _, p := range doc.Package {
		if p.Source == "" || p.Name == "" {
			continue
		}
		out = append(out, components.New("cargo", "", p.Name, p.Version))
	}
	return components.Dedupe(out), nil
}

// ComponentsFromManifest lists the dependencies a Cargo.toml declares, by
// the version requirement it asks for. Path-only dependencies are the
// workspace's own crates.
func ComponentsFromManifest(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Dependencies map[string]any `toml:"dependencies"`
	}
	if err := toml.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for name, v := range doc.Dependencies {
		switch spec := v.(type) {
		case string:
			out = append(out, components.New("cargo", "", name, spec))
		case map[string]any:
			if version, ok := spec["version"].(string); ok {
				out = append(out, components.New("cargo", "", name, version))
			}
		}
	}
	return components.Dedupe(out), nil
}
