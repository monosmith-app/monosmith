package pluginmarket

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"

	"github.com/monosmith-app/monosmith/apps/smith/pluginarchive"
	"github.com/monosmith-app/monosmith/apps/smith/plugininstall"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// ParseIndex reads and checks an index. It refuses another schema version, a
// plugin name or version that is not safe in a file name, a plugin with no
// version, a digest that is not 64 lowercase hex digits and an archive URL
// that is not https (http only for this machine). Unknown fields are ignored,
// so a newer index of the same schema still loads. Name and Version of each
// entry are filled from their keys.
func ParseIndex(data []byte) (Index, error) {
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return Index{}, fmt.Errorf("the marketplace index is not valid JSON: %w", err)
	}
	if idx.Schema != SchemaVersion {
		return Index{}, fmt.Errorf("the marketplace index has schema %d; this smith reads schema %d (update smith)", idx.Schema, SchemaVersion)
	}
	names := make([]string, 0, len(idx.Plugins))
	for name := range idx.Plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p := idx.Plugins[name]
		if err := pluginarchive.ValidatePluginName(name); err != nil {
			return Index{}, fmt.Errorf("marketplace index: %w", err)
		}
		if len(p.Versions) == 0 {
			return Index{}, fmt.Errorf("marketplace index: plugin %s lists no version", name)
		}
		p.Name = name
		for version, r := range p.Versions {
			if err := pluginarchive.ValidatePluginVersion(version); err != nil {
				return Index{}, fmt.Errorf("marketplace index: %s: %w", name, err)
			}
			if !digestPattern.MatchString(r.SHA256) {
				return Index{}, fmt.Errorf("marketplace index: %s@%s: sha256 must be 64 lowercase hex digits", name, version)
			}
			if err := plugininstall.CheckURL(r.URL); err != nil {
				return Index{}, fmt.Errorf("marketplace index: %s@%s: %w", name, version, err)
			}
			r.Version = version
			p.Versions[version] = r
		}
		idx.Plugins[name] = p
	}
	return idx, nil
}
