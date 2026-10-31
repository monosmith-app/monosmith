package sbom

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"go.yaml.in/yaml/v3"
)

// ComponentsFromLock lists the packages a pubspec.lock resolved. SDK
// packages (flutter), path packages and dev dependencies are not shipped from
// pub.dev and are left out.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Packages map[string]struct {
			Dependency string `yaml:"dependency"`
			Source     string `yaml:"source"`
			Version    string `yaml:"version"`
		} `yaml:"packages"`
	}
	if err := yaml.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for name, p := range doc.Packages {
		if p.Source == "sdk" || p.Source == "path" || p.Dependency == "direct dev" {
			continue
		}
		out = append(out, components.New("pub", "", name, p.Version))
	}
	return components.Dedupe(out), nil
}

// ComponentsFromManifest lists the dependencies of a pubspec.yaml, by the
// constraint it gives; SDK, path and git dependencies have none.
func ComponentsFromManifest(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Dependencies map[string]any `yaml:"dependencies"`
	}
	if err := yaml.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for name, v := range doc.Dependencies {
		switch spec := v.(type) {
		case string:
			out = append(out, components.New("pub", "", name, spec))
		case map[string]any:
			if version, ok := spec["version"].(string); ok {
				out = append(out, components.New("pub", "", name, version))
			}
		case nil:
			out = append(out, components.New("pub", "", name, ""))
		}
	}
	return components.Dedupe(out), nil
}
