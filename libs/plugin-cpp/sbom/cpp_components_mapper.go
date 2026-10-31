package sbom

import (
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ComponentsFromVcpkg lists the dependencies of a vcpkg.json. A version comes
// from an override, else from a "version>=" lower bound, else is empty (the
// registry baseline decides). vcpkg has no package URL type of its own, so
// they are generic packages.
func ComponentsFromVcpkg(manifest []byte) ([]protocol.Component, error) {
	var doc struct {
		Dependencies []json.RawMessage `json:"dependencies"`
		Overrides    []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"overrides"`
	}
	if err := json.Unmarshal(manifest, &doc); err != nil {
		return nil, err
	}
	pinned := map[string]string{}
	for _, o := range doc.Overrides {
		pinned[o.Name] = o.Version
	}
	var out []protocol.Component
	for _, raw := range doc.Dependencies {
		var name, version string
		var obj struct {
			Name    string `json:"name"`
			Version string `json:"version>="`
			Host    bool   `json:"host"`
		}
		if json.Unmarshal(raw, &name) != nil {
			if json.Unmarshal(raw, &obj) != nil || obj.Host {
				continue
			}
			name, version = obj.Name, obj.Version
		}
		if v, ok := pinned[name]; ok {
			version = v
		}
		if name != "" {
			out = append(out, components.New("generic", "", name, version))
		}
	}
	return components.Dedupe(out), nil
}

// ComponentsFromConan lists the [requires] of a conanfile.txt
// (name/version[@user/channel]).
func ComponentsFromConan(conanfile string) []protocol.Component {
	var out []protocol.Component
	inRequires := false
	for _, line := range strings.Split(conanfile, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "["):
			inRequires = line == "[requires]"
		case inRequires && line != "" && !strings.HasPrefix(line, "#"):
			ref, _, _ := strings.Cut(line, "@")
			name, version, _ := strings.Cut(ref, "/")
			out = append(out, components.New("conan", "", name, version))
		}
	}
	return components.Dedupe(out)
}
