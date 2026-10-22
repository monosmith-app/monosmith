package sbom

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ComponentsFromLock lists the packages an npm package-lock.json (v2 or v3)
// installs.
func ComponentsFromLock(lock []byte) ([]protocol.Component, error) {
	var doc struct {
		Packages map[string]struct {
			Version string `json:"version"`
			Link    bool   `json:"link"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for key, p := range doc.Packages {
		i := strings.LastIndex(key, "node_modules/")
		if key == "" || i < 0 || p.Link || p.Version == "" {
			continue
		}
		out = append(out, npm(key[i+len("node_modules/"):], p.Version))
	}
	return components.Dedupe(out), nil
}

// ComponentsFromManifest lists the dependencies a package.json declares, by
// the version range it asks for: the fallback when there is no lockfile.
func ComponentsFromManifest(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	var out []protocol.Component
	for name, v := range doc.Dependencies {
		out = append(out, npm(name, v))
	}
	return components.Dedupe(out), nil
}

// npm spells a scoped name as the package URL does: %40scope/name, with the
// scope's @ encoded and the slash kept.
func npm(name, version string) protocol.Component {
	c := components.New("npm", "", name, version)
	c.PURL = "pkg:npm/" + strings.Replace(name, "@", "%40", 1) + "@" + version
	return c
}
