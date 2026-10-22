package sbom

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ComponentsFromLock lists the runtime packages a composer.lock pins;
// packages-dev are tooling and are not shipped.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Packages []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for _, p := range doc.Packages {
		out = append(out, composer(p.Name, p.Version))
	}
	return components.Dedupe(out), nil
}

// ComponentsFromManifest lists the packages a composer.json requires, by
// constraint; php and its extensions are the platform, not packages.
func ComponentsFromManifest(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for name, constraint := range doc.Require {
		if name == "php" || strings.HasPrefix(name, "ext-") || strings.HasPrefix(name, "lib-") || !strings.Contains(name, "/") {
			continue
		}
		out = append(out, composer(name, constraint))
	}
	return components.Dedupe(out), nil
}

func composer(full, version string) protocol.Component {
	vendor, name, _ := strings.Cut(full, "/")
	return components.New("composer", vendor, name, version)
}
