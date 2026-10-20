package sbom

import (
	"encoding/json"
	"sort"
	"strings"
)

// NPMComponentsFromLock lists the packages an npm package-lock.json (v2 or
// v3) installs.
func NPMComponentsFromLock(lock []byte) ([]Component, error) {
	var doc struct {
		Packages map[string]struct {
			Version string `json:"version"`
			Link    bool   `json:"link"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(lock, &doc); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []Component
	for key, p := range doc.Packages {
		i := strings.LastIndex(key, "node_modules/")
		if key == "" || i < 0 || p.Link || p.Version == "" {
			continue
		}
		c := npmComponent(key[i+len("node_modules/"):], p.Version)
		if !seen[c.PURL] {
			seen[c.PURL] = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PURL < out[j].PURL })
	return out, nil
}

// NPMComponentsFromManifest lists the dependencies a package.json declares,
// by the version range it asks for: the fallback when there is no lockfile.
func NPMComponentsFromManifest(manifest []byte) ([]Component, error) {
	var doc struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	var out []Component
	for name, v := range doc.Dependencies {
		out = append(out, npmComponent(name, v))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PURL < out[j].PURL })
	return out, nil
}

func npmComponent(name, version string) Component {
	purl := "pkg:npm/" + strings.Replace(name, "@", "%40", 1) + "@" + version
	return Component{Name: name, Version: version, PURL: purl}
}
