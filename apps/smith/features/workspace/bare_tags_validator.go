package workspace

import (
	"fmt"
	"sort"
	"strings"
)

// validateBareTags checks bare_tags: it excludes tag_prefix, and only one
// project can use it, because bare tags of two projects could not be told
// apart.
func validateBareTags(cfg *Config) error {
	var bare []string
	for name, po := range cfg.Projects {
		if !po.BareTags {
			continue
		}
		if po.TagPrefix != "" {
			return fmt.Errorf("%s: [projects.%s] sets bare_tags and tag_prefix, which exclude each other", FileName, name)
		}
		bare = append(bare, name)
	}
	if len(bare) > 1 {
		sort.Strings(bare)
		return fmt.Errorf("%s: %s all set bare_tags, but only one project can tag releases with the bare version", FileName, strings.Join(bare, " and "))
	}
	return nil
}
