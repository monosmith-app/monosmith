package legacyinventory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Manifest is the part of a package.json the inventory reads.
type Manifest struct {
	Name                 string                             `json:"name"`
	Scripts              map[string]string                  `json:"scripts"`
	Dependencies         map[string]string                  `json:"dependencies"`
	DevDependencies      map[string]string                  `json:"devDependencies"`
	PeerDependencies     map[string]string                  `json:"peerDependencies"`
	OptionalDependencies map[string]string                  `json:"optionalDependencies"`
	PeerMeta             map[string]struct{ Optional bool } `json:"peerDependenciesMeta"`
	Bin                  json.RawMessage                    `json:"bin"`
}

// ReadManifest parses a package.json; ok is false when it is not valid JSON.
func ReadManifest(text string) (Manifest, bool) {
	var m Manifest
	if json.Unmarshal([]byte(text), &m) != nil {
		return Manifest{}, false
	}
	return m, true
}

// Declared lists the packages a manifest depends on, in every section.
func (m Manifest) Declared() []string {
	seen := map[string]bool{}
	for _, deps := range []map[string]string{m.Dependencies, m.DevDependencies, m.PeerDependencies, m.OptionalDependencies} {
		for name := range deps {
			seen[name] = true
		}
	}
	out := make([]string, 0, len(seen))
	for name := range seen {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// Bins are the command names the package installs.
func (m Manifest) Bins() []string {
	var one string
	if json.Unmarshal(m.Bin, &one) == nil && one != "" {
		return []string{m.Name[strings.LastIndex(m.Name, "/")+1:]}
	}
	var many map[string]string
	if json.Unmarshal(m.Bin, &many) != nil {
		return nil
	}
	out := make([]string, 0, len(many))
	for name := range many {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// RequiredPeers are the peer dependencies npm installs on its own: those not
// marked optional.
func (m Manifest) RequiredPeers() []string {
	var out []string
	for name := range m.PeerDependencies {
		if !m.PeerMeta[name].Optional {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// installed reads the manifest of a package in node_modules.
func (t *Tree) installed(name string) (Manifest, bool) {
	raw, err := os.ReadFile(filepath.Join(t.Modules, "node_modules", filepath.FromSlash(name), "package.json"))
	if err != nil {
		return Manifest{}, false
	}
	return ReadManifest(string(raw))
}
