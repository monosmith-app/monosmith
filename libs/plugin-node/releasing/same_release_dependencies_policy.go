package releasing

import (
	"encoding/json"
	"sort"
)

// SameReleaseOptionalDependencies lists the optionalDependencies of a manifest
// that are pinned to the version being published: the platform packages of one
// release (@acme/cli-linux-x64 beside @acme/cli). A range or another version
// is not part of this release and is left out.
//
// Parameters: manifest is the package.json; version is the version published.
// Returns: the package names, sorted; nil when there are none or the manifest
// does not parse.
func SameReleaseOptionalDependencies(manifest []byte, version string) []string {
	var m struct {
		Optional map[string]string `json:"optionalDependencies"`
	}
	if json.Unmarshal(manifest, &m) != nil {
		return nil
	}
	var names []string
	for name, v := range m.Optional {
		if v == version {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
